package resource

import (
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common/fileutil"
	"github.com/RosettaFlow/Carrier-Go/core"
	"github.com/RosettaFlow/Carrier-Go/core/rawdb"
	"github.com/RosettaFlow/Carrier-Go/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	defaultRefreshOrgResourceInterval = 30 * time.Second
)

type Manager struct {
	// TODO 这里需要一个 config <SlotUnit 的>
	dataCenter  core.CarrierDB // Low level persistent database to store final content.
	//eventCh                chan *libTypes.TaskEvent
	slotUnit *types.Slot
	remoteTableQueue     []*types.RemoteResourceTable
	remoteTableQueueMu  sync.RWMutex
	mockIdentityIdsFile  string
	mockIdentityIdsCache map[string]struct{}
}

func NewResourceManager(dataCenter core.CarrierDB, mockIdentityIdsFile string) *Manager {
	m := &Manager{
		dataCenter: dataCenter,
		remoteTableQueue:    make([]*types.RemoteResourceTable, 0),
		slotUnit:            types.DefaultSlotUnit, // TODO for test
		mockIdentityIdsFile: mockIdentityIdsFile,   //TODO for test
		mockIdentityIdsCache: make(map[string]struct{}, 0),
	}

	return m
}

func (m *Manager) loop() {
	refreshTicker := time.NewTicker(defaultRefreshOrgResourceInterval)
	for {
		select {
		case <-refreshTicker.C:
			if err := m.refreshOrgResourceTable(); nil != err {
				log.Errorf("Failed to refresh org resourceTables on loop, err: %s", err)
			}
		}
	}
}

func (m *Manager) Start() error {

	slotUnit, err := m.dataCenter.QueryNodeResourceSlotUnit()
	if nil != err {
		log.Warnf("Failed to load local slotUnit on resourceManager Start(), err: {%s}", err)
	} else {
		m.SetSlotUnit(slotUnit.Mem, slotUnit.Bandwidth, slotUnit.Processor)
	}

	// store slotUnit
	if err := m.dataCenter.StoreNodeResourceSlotUnit(m.slotUnit); nil != err {
		return err
	}
	// load remote org resource Tables
	remoteResources, err := m.dataCenter.QueryOrgResourceTables()
	if nil != err && err != rawdb.ErrNotFound {
		return err
	}
	if len(remoteResources) != 0 {
		m.remoteTableQueue = remoteResources
	} else {
		if err := m.refreshOrgResourceTable(); nil != err {
			log.Errorf("Failed to refresh org resourceTables on Start resourceManager, err: %s", err)
		}
	}

	// build mock identityIds cache
	if "" != m.mockIdentityIdsFile {
		var identityIdList []string
		if err := fileutil.LoadJSON(m.mockIdentityIdsFile, &identityIdList); err != nil {
			log.Errorf("Failed to load `--mock-identity-file` on Start resourceManager, file: {%s}, err: {%s}", m.mockIdentityIdsFile, err)
			return err
		}

		for _, iden := range identityIdList {
			m.mockIdentityIdsCache[iden] = struct{}{}
		}
	}



	go m.loop()
	log.Info("Started resourceManager ...")
	return nil
}

func (m *Manager) Stop() error {
	// store slotUnit
	if err := m.dataCenter.StoreNodeResourceSlotUnit(m.slotUnit); nil != err {
		return err
	}
	// store remote org resource Tables
	if err := m.dataCenter.StoreOrgResourceTables(m.remoteTableQueue); nil != err {
		return err
	}
	log.Infof("Stopped resource manager ...")
	return nil
}

func (m *Manager) SetSlotUnit(mem, b uint64, p uint32) {
	//m.slotUnit = &types.Slot{
	//	Mem:       mem,
	//	Processor: p,
	//	Bandwidth: b,
	//}
	m.slotUnit = types.DefaultSlotUnit // TODO for test
	//if len(m.localTables) != 0 {
	//	for _, re := range m.localTables {
	//		re.SetSlotUnit(m.slotUnit)
	//	}
	//}
}
func (m *Manager) GetSlotUnit() *types.Slot { return m.slotUnit }

func (m *Manager) UseSlot(nodeId string, slotCount uint32) error {
	table, err := m.GetLocalResourceTable(nodeId)
	if nil != err {
		return fmt.Errorf("No found the resource table of node: %s, %s", nodeId, err)
	}
	//if table.RemianSlot() < slotCount {
	//	return fmt.Errorf("Insufficient slotRemain {%s} less than need lock count {%s} slots of node: %s", table.RemianSlot(),slotCount , nodeId)
	//}
	if err := table.UseSlot(slotCount); nil != err {
		return err
	}
	return m.SetLocalResourceTable(table)
}
func (m *Manager) FreeSlot(nodeId string, slotCount uint32) error {
	table, err := m.GetLocalResourceTable(nodeId)
	if nil != err {
		return fmt.Errorf("No found the resource table of node: %s, %s", nodeId, err)
	}
	if err := table.FreeSlot(slotCount); nil != err {
		return err
	}
	return m.SetLocalResourceTable(table)
}

func (m *Manager) SetLocalResourceTable(table *types.LocalResourceTable) error {
	return m.dataCenter.StoreLocalResourceTable(table)
}
func (m *Manager) GetLocalResourceTable(nodeId string) (*types.LocalResourceTable, error) {
	return m.dataCenter.QueryLocalResourceTable(nodeId)
}
func (m *Manager) GetLocalResourceTables() ([]*types.LocalResourceTable, error) {
	return m.dataCenter.QueryLocalResourceTables()
}
func (m *Manager) DelLocalResourceTable(nodeId string) error {
	return m.dataCenter.RemoveLocalResourceTable(nodeId)
}
func (m *Manager) CleanLocalResourceTables() error {
	localResourceTableArr, err := m.dataCenter.QueryLocalResourceTables()
	if nil != err {
		return err
	}
	for _, table := range localResourceTableArr {
		if err := m.dataCenter.RemoveLocalResourceTable(table.GetNodeId()); nil != err {
			return err
		}
	}
	return nil
}

func (m *Manager) GetRemoteResourceTables() []*types.RemoteResourceTable {
	m.remoteTableQueueMu.RLock()
	defer m.remoteTableQueueMu.RUnlock()
	return m.remoteTableQueue
}
func (m *Manager) refreshOrgResourceTable() error {
	resources, err := m.dataCenter.GetResourceList()
	if nil != err {
		return err
	}

	remoteResourceArr := make([]*types.RemoteResourceTable, len(resources))

	for i, r := range resources {
		remoteResourceArr[i] = types.NewOrgResourceFromResource(r)
	}

	m.remoteTableQueueMu.Lock()
	defer m.remoteTableQueueMu.Unlock()

	m.remoteTableQueue = remoteResourceArr
	return nil
}

// TODO 有变更 RegisterNode mem  processor bandwidth 的 接口咩 ？？？
func (m *Manager) LockLocalResourceWithTask(jobNodeId string, needSlotCount uint64, task *types.Task) error {

	log.Infof("Start lock local resource with taskId {%s}, jobNodeId {%s}, slotCount {%d}", task.GetTaskId(), jobNodeId, needSlotCount)

	// Lock local resource (jobNode)
	if err := m.UseSlot(jobNodeId, uint32(needSlotCount)); nil != err {
		log.Errorf("Failed to lock internal power resource, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed to lock internal power resource, {%s}", err)
	}

	if err := m.dataCenter.StoreJobNodeRunningTaskId(jobNodeId, task.GetTaskId()); nil != err {

		m.FreeSlot(jobNodeId, uint32(needSlotCount))

		log.Errorf("Failed to store local taskId and jobNodeId index, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed to store local taskId and jobNodeId index, {%s}", err)
	}
	if err := m.dataCenter.StoreLocalTaskPowerUsed(types.NewLocalTaskPowerUsed(task.GetTaskId(), jobNodeId, needSlotCount)); nil != err {

		m.FreeSlot(jobNodeId, uint32(needSlotCount))
		m.dataCenter.RemoveJobNodeRunningTaskId(jobNodeId, task.GetTaskId())

		log.Errorf("Failed to store local taskId use jobNode slot, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed to store local taskId use jobNode slot, {%s}", err)
	}

	// 更新本地 resource 资源信息 [添加资源使用情况]
	jobNodeResource, err := m.dataCenter.GetLocalResource(jobNodeId)
	if nil != err {

		m.FreeSlot(jobNodeId, uint32(needSlotCount))
		m.dataCenter.RemoveJobNodeRunningTaskId(jobNodeId, task.GetTaskId())
		m.dataCenter.RemoveLocalTaskPowerUsed(task.GetTaskId())

		log.Errorf("Failed to query local jobNodeResource, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed to query local jobNodeResource, {%s}", err)
	}

	// 更新 本地 jobNodeResource 的资源使用信息
	usedMem := m.slotUnit.Mem * needSlotCount
	usedProcessor := m.slotUnit.Processor * uint32(needSlotCount)
	usedBandwidth := m.slotUnit.Bandwidth * needSlotCount

	jobNodeResource.GetData().UsedMem += usedMem
	jobNodeResource.GetData().UsedProcessor += uint32(usedProcessor)
	jobNodeResource.GetData().UsedBandwidth += usedBandwidth
	if err := m.dataCenter.InsertLocalResource(jobNodeResource); nil != err {

		m.FreeSlot(jobNodeId, uint32(needSlotCount))
		m.dataCenter.RemoveJobNodeRunningTaskId(jobNodeId, task.GetTaskId())
		m.dataCenter.RemoveLocalTaskPowerUsed(task.GetTaskId())

		log.Errorf("Failed to update local jobNodeResource, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed to update local jobNodeResource, {%s}", err)
	}

	// 还需要 将资源使用实况 实时上报给  dataCenter  [添加资源使用情况]
	if err := m.dataCenter.SyncPowerUsed(jobNodeResource); nil != err {
		log.Errorf("Failed to sync jobNodeResource to dataCenter, taskId: {%s}, jobNodeId: {%s}, usedSlotCount: {%s}, err: {%s}",
			task.GetTaskId(), jobNodeId, needSlotCount, err)
		return fmt.Errorf("failed tosync jobNodeResource to dataCenter, {%s}", err)
	}

	log.Infof("Finished lock local resource with taskId {%s}, jobNodeId {%s}, slotCount {%d}", task.GetTaskId(), jobNodeId, needSlotCount)
	return nil
}

// TODO 有变更 RegisterNode mem  processor bandwidth 的 接口咩 ？？？
func (m *Manager) UnLockLocalResourceWithTask(taskId string) error {

	localTaskPowerUsed, err := m.dataCenter.QueryLocalTaskPowerUsed(taskId)
	if nil != err {
		log.Errorf("Failed to query local taskId and jobNodeId index, err: %s", err)
		return fmt.Errorf("failed to query local taskId and jobNodeId index, err: %s", err)
	}

	jobNodeId := localTaskPowerUsed.GetNodeId()
	freeSlotUnitCount := localTaskPowerUsed.GetSlotCount()

	log.Infof("Start unlock local resource with taskId {%s}, jobNodeId {%s}, slotCount {%d}", taskId, jobNodeId, localTaskPowerUsed.GetSlotCount())

	// Lock local resource (jobNode)
	if err := m.FreeSlot(localTaskPowerUsed.GetNodeId(), uint32(freeSlotUnitCount)); nil != err {
		log.Errorf("Failed to unlock internal power resource, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to unlock internal power resource, {%s}", err)
	}

	if err := m.dataCenter.RemoveTaskEventList(taskId); nil != err {
		log.Errorf("Failed to remove local task event list, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to remove local task event list, {%s}", err)
	}

	if err := m.dataCenter.RemoveJobNodeRunningTaskId(jobNodeId, taskId); nil != err {
		log.Errorf("Failed to remove local taskId and jobNodeId index, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to remove local taskId and jobNodeId index, {%s}", err)
	}
	if err := m.dataCenter.RemoveLocalTaskPowerUsed(taskId); nil != err {
		log.Errorf("Failed to remove local taskId use jobNode slot, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to remove local taskId use jobNode slot, {%s}", err)
	}

	// 更新本地 resource 资源信息 [释放资源使用情况]
	jobNodeResource, err := m.dataCenter.GetLocalResource(jobNodeId)
	if nil != err {
		log.Errorf("Failed to query local jobNodeResource, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to query local jobNodeResource, {%s}", err)
	}

	// 更新 本地 jobNodeResource 的资源使用信息
	usedMem := m.slotUnit.Mem * freeSlotUnitCount
	usedProcessor := m.slotUnit.Processor * uint32(freeSlotUnitCount)
	usedBandwidth := m.slotUnit.Bandwidth * freeSlotUnitCount

	jobNodeResource.GetData().UsedMem -= usedMem
	jobNodeResource.GetData().UsedProcessor -= uint32(usedProcessor)
	jobNodeResource.GetData().UsedBandwidth -= usedBandwidth

	if err := m.dataCenter.InsertLocalResource(jobNodeResource); nil != err {
		log.Errorf("Failed to update local jobNodeResource, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed to update local jobNodeResource, {%s}", err)
	}

	// 还需要 将资源使用实况 实时上报给  dataCenter  [释放资源使用情况]
	if err := m.dataCenter.SyncPowerUsed(jobNodeResource); nil != err {
		log.Errorf("Failed to sync jobNodeResource to dataCenter, taskId: {%s}, jobNodeId: {%s}, freeSlotUnitCount: {%s}, err: {%s}",
			taskId, jobNodeId, freeSlotUnitCount, err)
		return fmt.Errorf("failed tosync jobNodeResource to dataCenter, {%s}", err)
	}

	log.Infof("Finished unlock local resource with taskId {%s}, jobNodeId {%s}, slotCount {%d}", taskId, localTaskPowerUsed.GetNodeId(), localTaskPowerUsed.GetSlotCount())
	return nil
}

func (m *Manager) ReleaseLocalResourceWithTask(logdesc, taskId string, option ReleaseResourceOption) {

	log.Debugf("Start ReleaseLocalResourceWithTask %s, taskId: {%s}, releaseOption: {%d}", logdesc, taskId, option)

	has, err := m.dataCenter.HasLocalTaskExecute(taskId)
	if nil != err {
		log.Errorf("Failed to query local task exec status with task %s, taskId: {%s}, err: {%s}", logdesc, taskId, err)
		return
	}

	if has {
		log.Debugf("The local task have been executing, don't `ReleaseLocalResourceWithTask` %s, taskId: {%s}, releaseOption: {%d}", logdesc, taskId, option)
		return
	}

	if option.IsUnlockLocalResorce() {
		log.Debugf("start unlock local resource with task %s, taskId: {%s}", logdesc, taskId)
		if err := m.UnLockLocalResourceWithTask(taskId); nil != err {
			log.Errorf("Failed to unlock local resource with task %s, taskId: {%s}, err: {%s}", logdesc, taskId, err)
		}
	}
	if option.IsRemoveLocalTask() {
		log.Debugf("start remove local task  %s, taskId: {%s}", logdesc, taskId)
		// 因为在 schedule 那边已经对 task 做了 StoreLocalTask
		if err := m.dataCenter.RemoveLocalTask(taskId); nil != err {
			log.Errorf("Failed to remove local task  %s, taskId: {%s}, err: {%s}", logdesc, taskId, err)
		}
	}
	if option.IsCleanTaskEvents() {
		log.Debugf("start clean event list of task  %s, taskId: {%s}", logdesc, taskId)
		if err := m.dataCenter.RemoveTaskEventList(taskId); nil != err {
			log.Errorf("Failed to clean event list of task  %s, taskId: {%s}, err: {%s}", logdesc, taskId, err)
		}
	}
}

func (m *Manager) IsMockIdentityId (identityId string) bool {
	if _, ok := m.mockIdentityIdsCache[identityId]; ok {
		return true
	}
	return false
}


/// ======================  v 2.0
func (m *Manager) GetDB() core.CarrierDB { return m.dataCenter }
