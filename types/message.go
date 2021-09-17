package types

import (
	"encoding/json"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common"
	"github.com/RosettaFlow/Carrier-Go/common/rlputil"
	"github.com/RosettaFlow/Carrier-Go/common/timeutils"
	pb "github.com/RosettaFlow/Carrier-Go/lib/api"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	"github.com/RosettaFlow/Carrier-Go/lib/types"
	libtypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"strings"
	"sync/atomic"
)

const (
	PREFIX_POWER_ID    = "power:"
	PREFIX_METADATA_ID = "metadata:"
	PREFIX_TASK_ID     = "task:"

	MSG_IDENTITY        = "identityMsg"
	MSG_IDENTITY_REVOKE = "identityRevokeMsg"
	MSG_POWER           = "powerMsg"
	MSG_POWER_REVOKE    = "powerRevokeMsg"
	MSG_METADATA        = "metaDataMsg"
	MSG_METADATA_REVOKE = "metaDataRevokeMsg"
	MSG_TASK            = "taskMsg"

	MSG_METADATAAUTHORITY       = "MetadataAuthorityMsg"
	MSG_METADATAAUTHORITYREVOKE = "MetadataAuthorityRevokeMsg"
)

type MessageType string

type Msg interface {
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
	String() string
	MsgType() string
}

// ------------------- identity -------------------

type IdentityMsg struct {
	organization *apicommonpb.Organization
	CreateAt     uint64 `json:"createAt"`
}

func NewIdentityMessageFromRequest(req *pb.ApplyIdentityJoinRequest) *IdentityMsg {
	return &IdentityMsg{
		organization: &apicommonpb.Organization{
			NodeName:   req.GetMember().GetNodeName(),
			NodeId:     req.GetMember().GetNodeId(),
			IdentityId: req.GetMember().GetIdentityId(),
		},
		CreateAt: uint64(timeutils.UnixMsec()),
	}
}

func (msg *IdentityMsg) ToDataCenter() *Identity {
	return NewIdentity(&libtypes.IdentityPB{
		NodeName:   msg.organization.NodeName,
		NodeId:     msg.organization.NodeId,
		IdentityId: msg.organization.IdentityId,
	})
}
func (msg *IdentityMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *IdentityMsg) Unmarshal(b []byte) error { return nil }
func (msg *IdentityMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *IdentityMsg) MsgType() string                            { return MSG_IDENTITY }
func (msg *IdentityMsg) GetOrganization() *apicommonpb.Organization { return msg.organization }
func (msg *IdentityMsg) GetOwnerName() string                       { return msg.organization.NodeName }
func (msg *IdentityMsg) GetOwnerNodeId() string                     { return msg.organization.NodeId }
func (msg *IdentityMsg) GetOwnerIdentityId() string                 { return msg.organization.IdentityId }
func (msg *IdentityMsg) GetCreateAt() uint64                        { return msg.CreateAt }
func (msg *IdentityMsg) SetOwnerNodeId(nodeId string)               { msg.organization.NodeId = nodeId }

type IdentityRevokeMsg struct {
	CreateAt uint64 `json:"createAt"`
}

func NewIdentityRevokeMessage() *IdentityRevokeMsg {
	return &IdentityRevokeMsg{
		CreateAt: uint64(timeutils.UnixMsec()),
	}
}

func (msg *IdentityRevokeMsg) GetCreateAt() uint64      { return msg.CreateAt }
func (msg *IdentityRevokeMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *IdentityRevokeMsg) Unmarshal(b []byte) error { return nil }
func (msg *IdentityRevokeMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *IdentityRevokeMsg) MsgType() string { return MSG_IDENTITY_REVOKE }

type IdentityMsgArr []*IdentityMsg
type IdentityRevokeMsgArr []*IdentityRevokeMsg

// Len returns the length of s.
func (s IdentityMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s IdentityMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s IdentityMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// Len returns the length of s.
func (s IdentityRevokeMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s IdentityRevokeMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s IdentityRevokeMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// ------------------- power -------------------

type PowerMsg struct {
	// This is only used when marshaling to JSON.
	PowerId   string `json:"powerId"`
	JobNodeId string `json:"jobNodeId"`
	CreateAt  uint64 `json:"createAt"`
	// caches
	hash atomic.Value
}

func NewPowerMessageFromRequest(req *pb.PublishPowerRequest) *PowerMsg {
	msg := &PowerMsg{
		JobNodeId: req.GetJobNodeId(),
		CreateAt:  uint64(timeutils.UnixMsec()),
	}
	msg.GenPowerId()
	return msg
}

func (msg *PowerMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *PowerMsg) Unmarshal(b []byte) error { return nil }
func (msg *PowerMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *PowerMsg) MsgType() string { return MSG_POWER }

func (msg *PowerMsg) GetJobNodeId() string { return msg.JobNodeId }
func (msg *PowerMsg) GetPowerId() string   { return msg.PowerId }
func (msg *PowerMsg) GetCreateAt() uint64  { return msg.CreateAt }
func (msg *PowerMsg) GenPowerId() string {
	if "" != msg.PowerId {
		return msg.PowerId
	}
	msg.PowerId = PREFIX_POWER_ID + msg.HashByCreateTime().Hex()
	return msg.PowerId
}

func (msg *PowerMsg) Hash() common.Hash {
	if hash := msg.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlputil.RlpHash([]interface{}{
		msg.JobNodeId,
	})
	msg.hash.Store(v)
	return v
}

func (msg *PowerMsg) HashByCreateTime() common.Hash {

	return rlputil.RlpHash([]interface{}{
		msg.JobNodeId,
		//msg.GetCreateAt,
		uint64(timeutils.UnixMsec()),
	})
}

type PowerRevokeMsg struct {
	PowerId  string `json:"powerId"`
	CreateAt uint64 `json:"createAt"`
}

func NewPowerRevokeMessageFromRequest(req *pb.RevokePowerRequest) *PowerRevokeMsg {
	return &PowerRevokeMsg{
		PowerId:  req.GetPowerId(),
		CreateAt: uint64(timeutils.UnixMsec()),
	}
}

func (msg *PowerRevokeMsg) GetPowerId() string       { return msg.PowerId }
func (msg *PowerRevokeMsg) GetCreateAt() uint64      { return msg.CreateAt }
func (msg *PowerRevokeMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *PowerRevokeMsg) Unmarshal(b []byte) error { return nil }
func (msg *PowerRevokeMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *PowerRevokeMsg) MsgType() string { return MSG_POWER_REVOKE }

type PowerMsgArr []*PowerMsg
type PowerRevokeMsgArr []*PowerRevokeMsg

// Len returns the length of s.
func (s PowerMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s PowerMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s PowerMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// Len returns the length of s.
func (s PowerRevokeMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s PowerRevokeMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s PowerRevokeMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// ------------------- metaData -------------------

type MetadataMsg struct {
	MetadataId      string                    `json:"metadataId"`
	MetadataSummary *libtypes.MetadataSummary `json:"metadataSummary"`
	ColumnMetas     []*types.MetadataColumn   `json:"columnMetas"`
	CreateAt        uint64                    `json:"createAt"`
	// caches
	hash atomic.Value
}

func NewMetadataMessageFromRequest(req *pb.PublishMetadataRequest) *MetadataMsg {
	metadataMsg := &MetadataMsg{
		MetadataSummary: &libtypes.MetadataSummary{
			MetadataId: req.GetInformation().GetMetadataSummary().GetMetadataId(),
			OriginId:   req.GetInformation().GetMetadataSummary().GetOriginId(),
			TableName:  req.GetInformation().GetMetadataSummary().GetTableName(),
			Desc:       req.GetInformation().GetMetadataSummary().GetDesc(),
			FilePath:   req.GetInformation().GetMetadataSummary().GetFilePath(),
			Rows:       req.GetInformation().GetMetadataSummary().GetRows(),
			Columns:    req.GetInformation().GetMetadataSummary().GetColumns(),
			Size_:      req.GetInformation().GetMetadataSummary().GetSize_(),
			FileType:   req.GetInformation().GetMetadataSummary().GetFileType(),
			HasTitle:   req.GetInformation().GetMetadataSummary().GetHasTitle(),
			State:      req.GetInformation().GetMetadataSummary().GetState(),
		},
		ColumnMetas: make([]*types.MetadataColumn, 0),
		CreateAt:    uint64(timeutils.UnixMsec()),
	}

	ColumnMetas := make([]*libtypes.MetadataColumn, len(req.GetInformation().GetMetadataColumns()))
	for i, v := range req.GetInformation().GetMetadataColumns() {
		ColumnMeta := &libtypes.MetadataColumn{
			CIndex:   v.GetCIndex(),
			CName:    v.GetCName(),
			CType:    v.GetCType(),
			CSize:    v.GetCSize(),
			CComment: v.GetCComment(),
		}
		ColumnMetas[i] = ColumnMeta
	}
	metadataMsg.ColumnMetas = ColumnMetas
	metadataMsg.GenMetadataId()

	return metadataMsg
}

func (msg *MetadataMsg) ToDataCenter(identity *apicommonpb.Organization) *Metadata {
	return NewMetadata(&libtypes.MetadataPB{
		MetadataId:      msg.GetMetadataId(),
		IdentityId:      identity.GetIdentityId(),
		NodeId:          identity.GetNodeId(),
		NodeName:        identity.GetNodeName(),
		DataId:          msg.GetMetadataId(),
		OriginId:        msg.GetOriginId(),
		TableName:       msg.GetTableName(),
		FilePath:        msg.GetFilePath(),
		FileType:        msg.GetFileType(),
		Desc:            msg.GetDesc(),
		Rows:            msg.GetRows(),
		Columns:         msg.GetColumns(),
		Size_:           msg.GetSize(),
		HasTitle:        msg.GetHasTitle(),
		MetadataColumns: msg.GetColumnMetas(),
		Industry:        msg.GetIndustry(),
		// the status of data, N means normal, D means deleted.
		DataStatus: apicommonpb.DataStatus_DataStatus_Normal,
		// metaData status, eg: create/release/revoke
		State: apicommonpb.MetadataState_MetadataState_Released,
	})
}
func (msg *MetadataMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *MetadataMsg) Unmarshal(b []byte) error { return nil }
func (msg *MetadataMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *MetadataMsg) MsgType() string { return MSG_METADATA }
func (msg *MetadataMsg) MetaDataSummary() *libtypes.MetadataSummary {
	return msg.MetadataSummary
}
func (msg *MetadataMsg) GetOriginId() string  { return msg.MetadataSummary.OriginId }
func (msg *MetadataMsg) GetTableName() string { return msg.MetadataSummary.TableName }
func (msg *MetadataMsg) GetDesc() string      { return msg.MetadataSummary.Desc }
func (msg *MetadataMsg) GetFilePath() string  { return msg.MetadataSummary.FilePath }
func (msg *MetadataMsg) GetRows() uint32      { return msg.MetadataSummary.Rows }
func (msg *MetadataMsg) GetColumns() uint32   { return msg.MetadataSummary.Columns }
func (msg *MetadataMsg) GetSize() uint64      { return msg.MetadataSummary.Size_ }
func (msg *MetadataMsg) GetFileType() apicommonpb.OriginFileType {
	return msg.MetadataSummary.FileType
}
func (msg *MetadataMsg) GetHasTitle() bool { return msg.MetadataSummary.HasTitle }
func (msg *MetadataMsg) GetState() apicommonpb.MetadataState {
	return msg.MetadataSummary.State
}
func (msg *MetadataMsg) GetColumnMetas() []*types.MetadataColumn {
	return msg.ColumnMetas
}
func (msg *MetadataMsg) GetIndustry() string   { return msg.MetadataSummary.Industry }
func (msg *MetadataMsg) GetCreateAt() uint64   { return msg.CreateAt }
func (msg *MetadataMsg) GetMetadataId() string { return msg.MetadataId }

func (msg *MetadataMsg) GenMetadataId() string {
	if "" != msg.MetadataId {
		return msg.MetadataId
	}
	msg.MetadataId = PREFIX_METADATA_ID + msg.HashByCreateTime().Hex()
	return msg.MetadataId
}
func (msg *MetadataMsg) Hash() common.Hash {
	if hash := msg.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlputil.RlpHash([]interface{}{
		msg.MetadataSummary,
		msg.ColumnMetas,
		msg.CreateAt,
	})
	msg.hash.Store(v)
	return v
}

func (msg *MetadataMsg) HashByCreateTime() common.Hash {
	return rlputil.RlpHash([]interface{}{
		msg.MetadataSummary.OriginId,
		uint64(timeutils.UnixMsec()),
	})
}

type MetadataRevokeMsg struct {
	MetadataId string `json:"metadataId"`
	CreateAt   uint64 `json:"createAt"`
}

func NewMetadataRevokeMessageFromRequest(req *pb.RevokeMetadataRequest) *MetadataRevokeMsg {
	return &MetadataRevokeMsg{
		MetadataId: req.GetMetadataId(),
		CreateAt:   uint64(timeutils.UnixMsec()),
	}
}

func (msg *MetadataRevokeMsg) GetMetadataId() string { return msg.MetadataId }
func (msg *MetadataRevokeMsg) GetCreateAat() uint64  { return msg.CreateAt }
func (msg *MetadataRevokeMsg) ToDataCenter(identity *apicommonpb.Organization) *Metadata {
	return NewMetadata(&libtypes.MetadataPB{
		MetadataId: msg.GetMetadataId(),
		IdentityId: identity.GetIdentityId(),
		NodeId:     identity.GetNodeId(),
		NodeName:   identity.GetNodeName(),
		DataId:     msg.GetMetadataId(),
		// the status of data, N means normal, D means deleted.
		DataStatus: apicommonpb.DataStatus_DataStatus_Deleted,
		// metaData status, eg: create/release/revoke
		State: apicommonpb.MetadataState_MetadataState_Revoked,
	})
}
func (msg *MetadataRevokeMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *MetadataRevokeMsg) Unmarshal(b []byte) error { return nil }
func (msg *MetadataRevokeMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *MetadataRevokeMsg) MsgType() string { return MSG_METADATA_REVOKE }

type MetadataMsgArr []*MetadataMsg
type MetadataRevokeMsgArr []*MetadataRevokeMsg

// Len returns the length of s.
func (s MetadataMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s MetadataMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s MetadataMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// Len returns the length of s.
func (s MetadataRevokeMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s MetadataRevokeMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s MetadataRevokeMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// ------------------- metadata authority apply -------------------

type MetadataAuthorityMsg struct {
	MetadataAuthId string                   `json:"metaDataAuthId"`
	User           string                   `json:"user"`
	UserType       apicommonpb.UserType     `json:"userType"`
	Auth           *types.MetadataAuthority `json:"auth"`
	Sign           []byte                   `json:"sign"`
	CreateAt       uint64                   `json:"createAt"`
	// caches
	hash atomic.Value
}

func NewMetadataAuthorityMessageFromRequest(req *pb.ApplyMetadataAuthorityRequest) *MetadataAuthorityMsg {
	metadataAuthorityMsg := &MetadataAuthorityMsg{
		User:     req.GetUser(),
		UserType: req.GetUserType(),
		Auth:     req.GetAuth(),
		Sign:     req.GetSign(),
		CreateAt: uint64(timeutils.UnixMsec()),
	}

	metadataAuthorityMsg.GenMetadataAuthId()

	return metadataAuthorityMsg
}

func (msg *MetadataAuthorityMsg) GetMetadataAuthId() string                      { return msg.MetadataAuthId }
func (msg *MetadataAuthorityMsg) GetUser() string                                { return msg.User }
func (msg *MetadataAuthorityMsg) GetUserType() apicommonpb.UserType              { return msg.UserType }
func (msg *MetadataAuthorityMsg) GetMetadataAuthority() *types.MetadataAuthority { return msg.Auth }
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityOwner() *apicommonpb.Organization {
	return msg.Auth.GetOwner()
}
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityOwnerIdentity() string {
	return msg.Auth.GetOwner().GetIdentityId()
}
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityOwnerNodeId() string {
	return msg.Auth.GetOwner().GetNodeId()
}
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityOwnerNodeName() string {
	return msg.Auth.GetOwner().GetNodeName()
}
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityMetadataId() string {
	return msg.Auth.GetMetadataId()
}
func (msg *MetadataAuthorityMsg) GetMetadataAuthorityUsageRule() *types.MetadataUsageRule {
	return msg.Auth.GetUsageRule()
}
func (msg *MetadataAuthorityMsg) GetSign() []byte     { return msg.Sign }
func (msg *MetadataAuthorityMsg) GetCreateAt() uint64 { return msg.CreateAt }

func (msg *MetadataAuthorityMsg) GenMetadataAuthId() string {
	if "" != msg.MetadataAuthId {
		return msg.MetadataAuthId
	}
	msg.MetadataAuthId = PREFIX_METADATA_ID + msg.HashByCreateTime().Hex()
	return msg.MetadataAuthId
}
func (msg *MetadataAuthorityMsg) Hash() common.Hash {
	if hash := msg.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlputil.RlpHash([]interface{}{
		msg.GetUser(),
		msg.GetUserType(),
		msg.GetMetadataAuthority(),
		msg.GetSign(),
		msg.CreateAt,
	})
	msg.hash.Store(v)
	return v
}

func (msg *MetadataAuthorityMsg) HashByCreateTime() common.Hash {
	return rlputil.RlpHash([]interface{}{
		msg.GetUser(),
		msg.GetUserType(),
		msg.GetMetadataAuthority().GetMetadataId(),
		msg.GetMetadataAuthority().GetUsageRule().GetStartAt(),
		msg.GetMetadataAuthority().GetUsageRule().GetEndAt(),
		msg.GetMetadataAuthority().GetUsageRule().GetTimes(),
		msg.GetSign(),
		uint64(timeutils.UnixMsec()),
	})
}

func (msg *MetadataAuthorityMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *MetadataAuthorityMsg) Unmarshal(b []byte) error { return nil }
func (msg *MetadataAuthorityMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *MetadataAuthorityMsg) MsgType() string { return MSG_METADATAAUTHORITY }

type MetadataAuthorityRevokeMsg struct {
	User           string
	UserType       apicommonpb.UserType
	MetadataAuthId string
	Sign           []byte
	CreateAt       uint64
}

func NewMetadataAuthorityRevokeMessageFromRequest(req *pb.RevokeMetadataAuthorityRequest) *MetadataAuthorityRevokeMsg {
	return &MetadataAuthorityRevokeMsg{
		User:           req.GetUser(),
		UserType:       req.GetUserType(),
		MetadataAuthId: req.GetMetadataAuthId(),
		Sign:           req.GetSign(),
		CreateAt:       uint64(timeutils.UnixMsec()),
	}
}

func (msg *MetadataAuthorityRevokeMsg) GetMetadataAuthId() string         { return msg.MetadataAuthId }
func (msg *MetadataAuthorityRevokeMsg) GetUser() string                   { return msg.User }
func (msg *MetadataAuthorityRevokeMsg) GetUserType() apicommonpb.UserType { return msg.UserType }
func (msg *MetadataAuthorityRevokeMsg) GetSign() []byte                   { return msg.Sign }
func (msg *MetadataAuthorityRevokeMsg) GetCreateAt() uint64               { return msg.CreateAt }

func (msg *MetadataAuthorityRevokeMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *MetadataAuthorityRevokeMsg) Unmarshal(b []byte) error { return nil }
func (msg *MetadataAuthorityRevokeMsg) String() string {
	result, err := json.Marshal(msg)
	if err != nil {
		return "Failed to generate string"
	}
	return string(result)
}
func (msg *MetadataAuthorityRevokeMsg) MsgType() string { return MSG_METADATAAUTHORITYREVOKE }

type MetadataAuthorityMsgArr []*MetadataAuthorityMsg
type MetadataAuthorityRevokeMsgArr []*MetadataAuthorityRevokeMsg

// Len returns the length of s.
func (s MetadataAuthorityMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s MetadataAuthorityMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s MetadataAuthorityMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// Len returns the length of s.
func (s MetadataAuthorityRevokeMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s MetadataAuthorityRevokeMsgArr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s MetadataAuthorityRevokeMsgArr) Less(i, j int) bool { return s[i].CreateAt < s[j].CreateAt }

// ------------------- task -------------------

type TaskBullet struct {
	TaskId  string
	Starve  bool
	Term    uint32
	Resched uint32
}

func NewTaskBullet(taskId string) *TaskBullet {
	return &TaskBullet{
		TaskId: taskId,
	}
}

func (b *TaskBullet) IncreaseResched() { b.Resched++ }
func (b *TaskBullet) DecreaseResched() {
	if b.Resched > 0 {
		b.Resched--
	}
}
func (b *TaskBullet) IncreaseTerm() { b.Term++ }
func (b *TaskBullet) DecreaseTerm() {
	if b.Term > 0 {
		b.Term--
	}
}

type TaskBullets []*TaskBullet

func (h TaskBullets) Len() int           { return len(h) }
func (h TaskBullets) Less(i, j int) bool { return h[i].Term > h[j].Term } // term:  a.3 > c.2 > b.1,  So order is: a c b
func (h TaskBullets) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *TaskBullets) Push(x interface{}) {
	*h = append(*h, x.(*TaskBullet))
}

func (h *TaskBullets) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func (h *TaskBullets) IncreaseTerm() {
	for i := range *h {
		(*h)[i].IncreaseTerm()
	}
}
func (h *TaskBullets) DecreaseTerm() {
	for i := range *h {
		(*h)[i].DecreaseTerm()
	}
}

type TaskMsg struct {
	Data          *Task
	PowerPartyIds []string `json:"powerPartyIds"`
	// caches
	hash atomic.Value
}

func NewTaskMessageFromRequest(req *pb.PublishTaskDeclareRequest) *TaskMsg {
	return &TaskMsg{
		PowerPartyIds: req.GetPowerPartyIds(),
		Data: NewTask(&libtypes.TaskPB{
			TaskId:       "",
			TaskName:     req.GetTaskName(),
			PartyId:      req.GetSender().GetPartyId(),
			IdentityId:   req.GetSender().GetIdentityId(),
			NodeId:       req.GetSender().GetNodeId(),
			NodeName:     req.GetSender().GetNodeName(),
			DataId:       "",
			DataStatus:   apicommonpb.DataStatus_DataStatus_Normal,
			State:        apicommonpb.TaskState_TaskState_Pending,
			Reason:       "",
			EventCount:   0,
			Desc:         "",
			CreateAt:     uint64(timeutils.UnixMsec()),
			EndAt:        0,
			StartAt:      0,
			AlgoSupplier: req.GetAlgoSupplier(),
			//DataSuppliers:
			PowerSuppliers:        make([]*libtypes.TaskPowerSupplier, 0),
			Receivers:             req.GetReceivers(),
			OperationCost:         req.GetOperationCost(),
			CalculateContractCode: req.GetCalculateContractCode(),
			DataSplitContractCode: req.GetDataSplitContractCode(),
			ContractExtraParams:   req.GetContractExtraParams(),
		}),
	}
}
func ConvertTaskMsgToTaskWithPowers(task *Task, powers []*libtypes.TaskPowerSupplier) *Task {
	task.SetResourceSupplierArr(powers)
	return task
}

type TaskMsgArr []*TaskMsg

func (msg *TaskMsg) Marshal() ([]byte, error) { return nil, nil }
func (msg *TaskMsg) Unmarshal(b []byte) error { return nil }
func (msg *TaskMsg) String() string {
	return fmt.Sprintf(`{"taskId": %s, "powerPartyIds": %s, "task": %s}`,
		msg.Data.GetTaskId(), "["+strings.Join(msg.PowerPartyIds, ",")+"]", msg.Data.GetTaskData().String())
}
func (msg *TaskMsg) MsgType() string { return MSG_TASK }

func (msg *TaskMsg) GetUserType() apicommonpb.UserType { return msg.Data.GetTaskData().GetUserType() }
func (msg *TaskMsg) GetUser() string                   { return msg.Data.GetTaskData().GetUser() }
func (msg *TaskMsg) GetSender() *apicommonpb.TaskOrganization {
	return &apicommonpb.TaskOrganization{
		PartyId:    msg.Data.GetTaskData().GetPartyId(),
		NodeName:   msg.Data.GetTaskData().GetNodeName(),
		NodeId:     msg.Data.GetTaskData().GetNodeId(),
		IdentityId: msg.Data.GetTaskData().GetIdentityId(),
	}
}
func (msg *TaskMsg) GetSenderName() string       { return msg.Data.GetTaskData().GetNodeName() }
func (msg *TaskMsg) GetSenderNodeId() string     { return msg.Data.GetTaskData().GetNodeId() }
func (msg *TaskMsg) GetSenderIdentityId() string { return msg.Data.GetTaskData().GetIdentityId() }
func (msg *TaskMsg) GetSenderPartyId() string    { return msg.Data.GetTaskData().GetPartyId() }
func (msg *TaskMsg) GetTaskId() string           { return msg.Data.GetTaskData().GetTaskId() }
func (msg *TaskMsg) GetTaskName() string         { return msg.Data.GetTaskData().GetTaskName() }
func (msg *TaskMsg) GetAlgoSupplier() *apicommonpb.TaskOrganization {
	return &apicommonpb.TaskOrganization{
		PartyId:    msg.Data.GetTaskData().GetAlgoSupplier().GetPartyId(),
		NodeName:   msg.Data.GetTaskData().GetAlgoSupplier().GetNodeName(),
		NodeId:     msg.Data.GetTaskData().GetAlgoSupplier().GetNodeId(),
		IdentityId: msg.Data.GetTaskData().GetAlgoSupplier().GetIdentityId(),
	}
}
func (msg *TaskMsg) GetTaskMetadataSuppliers() []*apicommonpb.TaskOrganization {

	partners := make([]*apicommonpb.TaskOrganization, len(msg.Data.GetTaskData().GetDataSuppliers()))
	for i, v := range msg.Data.GetTaskData().GetDataSuppliers() {

		partners[i] = &apicommonpb.TaskOrganization{
			PartyId:    v.GetOrganization().GetPartyId(),
			NodeName:   v.GetOrganization().GetNodeName(),
			NodeId:     v.GetOrganization().GetNodeId(),
			IdentityId: v.GetOrganization().GetIdentityId(),
		}
	}
	return partners
}
func (msg *TaskMsg) GetTaskMetadataSupplierDatas() []*libtypes.TaskDataSupplier {

	return msg.Data.GetTaskData().GetDataSuppliers()
}

func (msg *TaskMsg) GetTaskResourceSuppliers() []*apicommonpb.TaskOrganization {
	powers := make([]*apicommonpb.TaskOrganization, len(msg.Data.GetTaskData().GetPowerSuppliers()))
	for i, v := range msg.Data.GetTaskData().GetPowerSuppliers() {
		powers[i] = &apicommonpb.TaskOrganization{
			PartyId:    v.GetOrganization().GetPartyId(),
			NodeName:   v.GetOrganization().GetNodeName(),
			NodeId:     v.GetOrganization().GetNodeId(),
			IdentityId: v.GetOrganization().GetIdentityId(),
		}
	}
	return powers
}
func (msg *TaskMsg) GetTaskResourceSupplierDatas() []*libtypes.TaskPowerSupplier {
	return msg.Data.GetTaskData().GetPowerSuppliers()
}
func (msg *TaskMsg) GetPowerPartyIds() []string { return msg.PowerPartyIds }
func (msg *TaskMsg) GetReceivers() []*apicommonpb.TaskOrganization {
	return msg.Data.GetTaskData().GetReceivers()
}

func (msg *TaskMsg) GetCalculateContractCode() string {
	return msg.Data.GetTaskData().GetCalculateContractCode()
}
func (msg *TaskMsg) GetDataSplitContractCode() string {
	return msg.Data.GetTaskData().GetDataSplitContractCode()
}
func (msg *TaskMsg) GetContractExtraParams() string {
	return msg.Data.GetTaskData().GetContractExtraParams()
}
func (msg *TaskMsg) GetOperationCost() *apicommonpb.TaskResourceCostDeclare {
	return msg.Data.GetTaskData().GetOperationCost()
}
func (msg *TaskMsg) GetCreateAt() uint64 { return msg.Data.GetTaskData().GetCreateAt() }
func (msg *TaskMsg) GenTaskId() string {
	if "" != msg.Data.GetTaskId() {
		return msg.Data.GetTaskId()
	}
	msg.Data.GetTaskData().TaskId = PREFIX_TASK_ID + msg.HashByCreateTime().Hex()
	return msg.Data.GetTaskId()
}
func (msg *TaskMsg) Hash() common.Hash {
	if hash := msg.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlputil.RlpHash(msg.Data.GetTaskData())
	msg.hash.Store(v)
	return v
}

func (msg *TaskMsg) HashByCreateTime() common.Hash {
	return rlputil.RlpHash([]interface{}{
		msg.Data.GetTaskData().GetIdentityId(),
		msg.Data.GetTaskData().GetPartyId(),
		msg.Data.GetTaskData().GetTaskName(),
		//msg.Data.GetTaskData().GetCreateAt,
		uint64(timeutils.UnixMsec()),
	})
}

// Len returns the length of s.
func (s TaskMsgArr) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s TaskMsgArr) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TaskMsgArr) Less(i, j int) bool {
	return s[i].Data.GetTaskData().GetCreateAt() < s[j].Data.GetTaskData().GetCreateAt()
}

/**
Example:
{
  "party_id": "p0",
  "data_party": {
      "input_file": "../data/bank_predict_data.csv",
       "key_column": "CLIENT_ID",
       "selected_columns": ["col1", "col2"]
    },
  "dynamic_parameter": {
    "model_restore_party": "p0",
    "train_task_id": "task_id"
  }
}

or:

{
  "party_id": "p0",
  "data_party": {
    "input_file": "../data/bank_train_data.csv",
    "key_column": "CLIENT_ID",
    "selected_columns": ["col1", "col2"]
  },
  "dynamic_parameter": {
    "label_owner": "p0",
    "label_column_name": "Y",
    "algorithm_parameter": {
      "epochs": 10,
      "batch_size": 256,
      "learning_rate": 0.1
    }
  }
}
*/
type FighterTaskReadyGoReqContractCfg struct {
	PartyId   string `json:"party_id"`
	DataParty struct {
		InputFile       string   `json:"input_file"`
		KeyColumn       string   `json:"key_column"`
		SelectedColumns []string `json:"selected_columns"`
	} `json:"data_party"`
	DynamicParameter map[string]interface{} `json:"dynamic_parameter"`
}
