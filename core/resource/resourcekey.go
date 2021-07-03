package resource

import "github.com/RosettaFlow/Carrier-Go/common"

var (
	nodeResourceKeyPrefix = []byte("NodeResourceKey:")
	nodeResourceIdListKey = []byte("nodeResourceIdListKey:")
)


// nodeResourceKey = NodeResourceKeyPrefix + jobNodeId
func GetNodeResourceKey(jobNodeId string) []byte {
	return append(nodeResourceKeyPrefix, common.Hex2Bytes(jobNodeId)...)
}
func GetNodeResourceIdListKey() []byte {
	return nodeResourceIdListKey
}