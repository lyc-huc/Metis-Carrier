package backend

import (
	pb "github.com/RosettaFlow/Carrier-Go/lib/api"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	libtypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"github.com/RosettaFlow/Carrier-Go/types"
)

type Backend interface {
	SendMsg(msg types.Msg) error

	// system (the yarn node self info)
	GetNodeInfo() (*pb.YarnNodeInfo, error)
	GetRegisteredPeers(nodeType pb.NodeType) ([]*pb.YarnRegisteredPeer, error)

	// local node resource api

	SetSeedNode(seed *pb.SeedPeer) (pb.ConnState, error)
	DeleteSeedNode(addr string) error
	GetSeedNodeList() ([]*pb.SeedPeer, error)
	SetRegisterNode(typ pb.RegisteredNodeType, node *pb.YarnRegisteredPeerDetail) (pb.ConnState, error)
	UpdateRegisterNode(typ pb.RegisteredNodeType, node *pb.YarnRegisteredPeerDetail) (pb.ConnState, error)
	DeleteRegisterNode(typ pb.RegisteredNodeType, id string) error
	GetRegisterNode(typ pb.RegisteredNodeType, id string) (*pb.YarnRegisteredPeerDetail, error)
	GetRegisterNodeList(typ pb.RegisteredNodeType) ([]*pb.YarnRegisteredPeerDetail, error)

	SendTaskEvent(event *libtypes.TaskEvent) error

	// v 2.0
	ReportTaskResourceUsage(nodeType pb.NodeType, ip, port string, usage *types.TaskResuorceUsage) error


	// metadata api
	IsInternalMetadata(metadataId string) (bool, error)
	GetMetadataDetail(identityId, metadataId string) (*types.Metadata, error)
	GetGlobalMetadataDetailList(lastUpdate uint64) ([]*pb.GetGlobalMetadataDetailResponse, error)
	GetLocalMetadataDetailList(lastUpdate uint64) ([]*pb.GetLocalMetadataDetailResponse, error)
	GetMetadataUsedTaskIdList(identityId, metadataId string) ([]string, error)

	// metadataAuthority api

	AuditMetadataAuthority(audit *types.MetadataAuthAudit) (apicommonpb.AuditMetadataOption, error)
	GetLocalMetadataAuthorityList() (types.MetadataAuthArray, error)
	GetGlobalMetadataAuthorityList(lastUpdate uint64) (types.MetadataAuthArray, error)
	HasValidMetadataAuth(userType apicommonpb.UserType, user, identityId, metadataId string) (bool, error)

	// power api
	GetGlobalPowerSummaryList(lastUpdate uint64) ([]*pb.GetGlobalPowerSummaryResponse, error)
	GetGlobalPowerDetailList(lastUpdate uint64) ([]*pb.GetGlobalPowerDetailResponse, error)
	GetLocalPowerDetailList(lastUpdate uint64) ([]*pb.GetLocalPowerDetailResponse, error)

	// identity api

	GetNodeIdentity() (*types.Identity, error)
	GetIdentityList(lastUpdate uint64) ([]*types.Identity, error)

	// task api
	GetLocalTask(taskId string) (*pb.TaskDetailShow, error)
	GetTaskDetailList(lastUpdate uint64) ([]*pb.TaskDetailShow, error)
	GetTaskEventList(taskId string) ([]*pb.TaskEventShow, error)
	GetTaskEventListByTaskIds(taskIds []string) ([]*pb.TaskEventShow, error)
	HasLocalTask () (bool, error)

	// about DataResourceTable

	StoreDataResourceTable(dataResourceTable *types.DataResourceTable) error
	StoreDataResourceTables(dataResourceTables []*types.DataResourceTable) error
	RemoveDataResourceTable(nodeId string) error
	QueryDataResourceTable(nodeId string) (*types.DataResourceTable, error)
	QueryDataResourceTables() ([]*types.DataResourceTable, error)

	// about DataResourceFileUpload

	StoreDataResourceFileUpload(dataResourceDataUsed *types.DataResourceFileUpload) error
	StoreDataResourceFileUploads(dataResourceDataUseds []*types.DataResourceFileUpload) error
	RemoveDataResourceFileUpload(originId string) error
	QueryDataResourceFileUpload(originId string) (*types.DataResourceFileUpload, error)
	QueryDataResourceFileUploads() ([]*types.DataResourceFileUpload, error)

	// about task result file

	StoreTaskUpResultFile(turf *types.TaskUpResultFile) error
	QueryTaskUpResultFile(taskId string) (*types.TaskUpResultFile, error)
	RemoveTaskUpResultFile(taskId string) error
	StoreTaskResultFileSummary(taskId, originId, filePath, dataNodeId string) error
	QueryTaskResultFileSummary (taskId string) (*types.TaskResultFileSummary, error)
	QueryTaskResultFileSummaryList () (types.TaskResultFileSummaryArr, error)
}
