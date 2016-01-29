/*Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package skyring

import (
	"fmt"
	"github.com/skyrings/skyring/conf"
	"github.com/skyrings/skyring/db"
	"github.com/skyrings/skyring/models"
	"github.com/skyrings/skyring/tools/lock"
	"github.com/skyrings/skyring/tools/logger"
	"github.com/skyrings/skyring/tools/uuid"
	"github.com/skyrings/skyring/utils"
	"gopkg.in/mgo.v2/bson"
)

func lockNode(nodeId uuid.UUID, hostname string, operation string) (*lock.AppLock, error) {
	//lock the node
	locks := make(map[uuid.UUID]string)
	if nodeId.IsZero() {
		//Generate temporary UUID from hostname for the node
		//for locking as the UUID is not available at this point
		id, err := uuid.Parse(util.Md5FromString(hostname))
		if err != nil {
			logger.Get().Error(fmt.Sprintf("Unable to create the UUID for locking for host: %s:", hostname), err)
			return nil, err
		}
		nodeId = *id
	}
	locks[nodeId] = fmt.Sprintf("%s : %s", operation, hostname)
	appLock := lock.NewAppLock(locks)
	if err := GetApp().GetLockManager().AcquireLock(*appLock); err != nil {
		return nil, err
	}
	return appLock, nil
}

func lockNodes(nodes models.Nodes, operation string) (*lock.AppLock, error) {
	locks := make(map[uuid.UUID]string)
	for _, node := range nodes {
		locks[node.NodeId] = fmt.Sprintf("%s : %s", operation, node.Hostname)
	}
	appLock := lock.NewAppLock(locks)
	if err := GetApp().GetLockManager().AcquireLock(*appLock); err != nil {
		return nil, err
	}
	return appLock, nil
}

func getClusterNodesById(cluster_id *uuid.UUID) (models.Nodes, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	collection := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	var nodes models.Nodes
	if err := collection.Find(bson.M{"clusterid": *cluster_id}).All(&nodes); err != nil {
		return nil, err
	}
	return nodes, nil

}

func getClusterNodesFromRequest(clusterNodes []models.ClusterNode) (models.Nodes, error) {
	sessionCopy := db.GetDatastore().Copy()
	defer sessionCopy.Close()

	var nodes models.Nodes
	coll := sessionCopy.DB(conf.SystemConfig.DBConfig.Database).C(models.COLL_NAME_STORAGE_NODES)
	for _, clusterNode := range clusterNodes {
		uuid, err := uuid.Parse(clusterNode.NodeId)
		if err != nil {
			return nodes, err
		}
		var node models.Node
		if err := coll.Find(bson.M{"nodeid": *uuid}).One(&node); err != nil {
			return nodes, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil

}