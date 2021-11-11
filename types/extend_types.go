package types

import (
	pb "github.com/RosettaFlow/Carrier-Go/lib/api"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	libtypes "github.com/RosettaFlow/Carrier-Go/lib/types"
)

func NewTaskDetailShowFromTaskData(input *Task) *pb.TaskDetailShow {
	taskData := input.GetTaskData()
	detailShow := &pb.TaskDetailShow{
		TaskId:   taskData.GetTaskId(),
		TaskName: taskData.GetTaskName(),
		UserType: taskData.GetUserType(),
		User:     taskData.GetUser(),
		Sender: &apicommonpb.TaskOrganization{
			PartyId:    input.GetTaskSender().GetPartyId(),
			NodeName:   input.GetTaskSender().GetNodeName(),
			NodeId:     input.GetTaskSender().GetNodeId(),
			IdentityId: input.GetTaskSender().GetIdentityId(),
		},
		AlgoSupplier: &apicommonpb.TaskOrganization{
			PartyId:    input.GetTaskData().GetAlgoSupplier().GetPartyId(),
			NodeName:   input.GetTaskData().GetAlgoSupplier().GetNodeName(),
			NodeId:     input.GetTaskData().GetAlgoSupplier().GetNodeId(),
			IdentityId: input.GetTaskData().GetAlgoSupplier().GetIdentityId(),
		},
		DataSuppliers:  make([]*pb.TaskDataSupplierShow, 0, len(taskData.GetDataSuppliers())),
		PowerSuppliers: make([]*pb.TaskPowerSupplierShow, 0, len(taskData.GetPowerSuppliers())),
		Receivers:      taskData.GetReceivers(),
		CreateAt:       taskData.GetCreateAt(),
		StartAt:        taskData.GetStartAt(),
		EndAt:          taskData.GetEndAt(),
		State:          taskData.GetState(),
		OperationCost: &apicommonpb.TaskResourceCostDeclare{
			Processor: taskData.GetOperationCost().GetProcessor(),
			Memory:    taskData.GetOperationCost().GetMemory(),
			Bandwidth: taskData.GetOperationCost().GetBandwidth(),
			Duration:  taskData.GetOperationCost().GetDuration(),
		},
	}

	// DataSupplier
	for _, dataSupplier := range taskData.GetDataSuppliers() {
		detailShow.DataSuppliers = append(detailShow.DataSuppliers, &pb.TaskDataSupplierShow{
			Organization: &apicommonpb.TaskOrganization{
				PartyId:    dataSupplier.GetOrganization().GetPartyId(),
				NodeName:   dataSupplier.GetOrganization().GetNodeName(),
				NodeId:     dataSupplier.GetOrganization().GetNodeId(),
				IdentityId: dataSupplier.GetOrganization().GetIdentityId(),
			},
			MetadataId:   dataSupplier.GetMetadataId(),
			MetadataName: dataSupplier.GetMetadataName(),
		})
	}
	// powerSupplier
	for _, data := range taskData.GetPowerSuppliers() {
		detailShow.PowerSuppliers = append(detailShow.PowerSuppliers, &pb.TaskPowerSupplierShow{
			Organization: &apicommonpb.TaskOrganization{
				PartyId:    data.GetOrganization().GetPartyId(),
				NodeName:   data.GetOrganization().GetNodeName(),
				NodeId:     data.GetOrganization().GetNodeId(),
				IdentityId: data.GetOrganization().GetIdentityId(),
			},
			PowerInfo: &libtypes.ResourceUsageOverview{
				TotalMem:       data.GetResourceUsedOverview().GetTotalMem(),
				UsedMem:        data.GetResourceUsedOverview().GetUsedMem(),
				TotalProcessor: data.GetResourceUsedOverview().GetTotalProcessor(),
				UsedProcessor:  data.GetResourceUsedOverview().GetUsedProcessor(),
				TotalBandwidth: data.GetResourceUsedOverview().GetTotalBandwidth(),
				UsedBandwidth:  data.GetResourceUsedOverview().GetUsedBandwidth(),
				TotalDisk:      data.GetResourceUsedOverview().GetTotalDisk(),
				UsedDisk:       data.GetResourceUsedOverview().GetUsedDisk(),
			},
		})
	}
	return detailShow
}

func NewTaskEventFromAPIEvent(input []*libtypes.TaskEvent) []*pb.TaskEventShow {
	result := make([]*pb.TaskEventShow, 0, len(input))
	for _, event := range input {
		result = append(result, &pb.TaskEventShow{
			TaskId:   event.GetTaskId(),
			Type:     event.GetType(),
			CreateAt: event.GetCreateAt(),
			Content:  event.GetContent(),
			Owner: &apicommonpb.Organization{
				IdentityId: event.GetIdentityId(),
			},
			PartyId: event.GetPartyId(),
		})
	}
	return result
}

func NewGlobalMetadataInfoFromMetadata(input *Metadata) *pb.GetGlobalMetadataDetailResponse {
	response := &pb.GetGlobalMetadataDetailResponse{
		Owner: &apicommonpb.Organization{
			NodeName:   input.data.GetNodeName(),
			NodeId:     input.data.GetNodeId(),
			IdentityId: input.data.GetIdentityId(),
		},
		Information: &libtypes.MetadataDetail{
			MetadataSummary: &libtypes.MetadataSummary{
				MetadataId: input.GetData().GetMetadataId(),
				OriginId:   input.GetData().GetOriginId(),
				TableName:  input.GetData().GetTableName(),
				Desc:       input.GetData().GetDesc(),
				FilePath:   input.GetData().GetFilePath(),
				Rows:       input.GetData().GetRows(),
				Columns:    input.GetData().GetColumns(),
				Size_:      input.GetData().GetSize_(),
				FileType:   input.GetData().GetFileType(),
				HasTitle:   input.GetData().GetHasTitle(),
				Industry:   input.GetData().GetIndustry(),
				State:      input.GetData().GetState(),
				PublishAt:  input.GetData().GetPublishAt(),
				UpdateAt:   input.GetData().GetUpdateAt(),
			},
			MetadataColumns: input.GetData().GetMetadataColumns(),
		},
	}
	return response
}

func NewLocalMetadataInfoFromMetadata(isInternal bool, input *Metadata) *pb.GetLocalMetadataDetailResponse {
	response := &pb.GetLocalMetadataDetailResponse{
		Owner: &apicommonpb.Organization{
			NodeName:   input.data.GetNodeName(),
			NodeId:     input.data.GetNodeId(),
			IdentityId: input.data.GetIdentityId(),
		},
		Information: &libtypes.MetadataDetail{
			MetadataSummary: &libtypes.MetadataSummary{
				MetadataId: input.GetData().GetDataId(),
				OriginId:   input.GetData().GetOriginId(),
				TableName:  input.GetData().GetTableName(),
				Desc:       input.GetData().GetDesc(),
				FilePath:   input.GetData().GetFilePath(),
				Rows:       input.GetData().GetRows(),
				Columns:    input.GetData().GetColumns(),
				Size_:      input.GetData().GetSize_(),
				FileType:   input.GetData().GetFileType(),
				HasTitle:   input.GetData().GetHasTitle(),
				Industry:   input.GetData().GetIndustry(),
				State:      input.GetData().GetState(),
				PublishAt:  input.GetData().GetPublishAt(),
				UpdateAt:   input.GetData().GetUpdateAt(),
			},
			MetadataColumns: input.GetData().GetMetadataColumns(),
		},
		IsInternal: isInternal,
	}
	return response
}

func NewGlobalMetadataInfoArrayFromMetadataArray(input MetadataArray) []*pb.GetGlobalMetadataDetailResponse {
	result := make([]*pb.GetGlobalMetadataDetailResponse, 0, input.Len())
	for _, metadata := range input {
		if metadata == nil {
			continue
		}
		result = append(result, NewGlobalMetadataInfoFromMetadata(metadata))
	}
	return result
}

func NewLocalMetadataInfoArrayFromMetadataArray(internalArr, publishArr MetadataArray) []*pb.GetLocalMetadataDetailResponse {
	result := make([]*pb.GetLocalMetadataDetailResponse, 0, internalArr.Len()+publishArr.Len())

	for _, metadata := range internalArr {
		if metadata == nil {
			continue
		}
		result = append(result, NewLocalMetadataInfoFromMetadata(true, metadata))
	}

	for _, metadata := range publishArr {
		if metadata == nil {
			continue
		}
		result = append(result, NewLocalMetadataInfoFromMetadata(false, metadata))
	}

	return result
}
