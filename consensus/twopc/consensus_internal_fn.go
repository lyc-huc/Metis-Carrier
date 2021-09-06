package twopc

import (
	"context"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common"
	"github.com/RosettaFlow/Carrier-Go/common/timeutils"
	ctypes "github.com/RosettaFlow/Carrier-Go/consensus/twopc/types"
	"github.com/RosettaFlow/Carrier-Go/core/evengine"
	"github.com/RosettaFlow/Carrier-Go/core/resource"
	"github.com/RosettaFlow/Carrier-Go/handler"
	apipb "github.com/RosettaFlow/Carrier-Go/lib/common"
	pb "github.com/RosettaFlow/Carrier-Go/lib/consensus/twopc"
	libTypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"github.com/RosettaFlow/Carrier-Go/p2p"
	"github.com/RosettaFlow/Carrier-Go/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"strings"
	"sync"
)




func (t *TwoPC) isProposalTask(taskId string) bool {
	t.proposalTaskLock.RLock()
	_, ok := t.proposalTaskCache[taskId]
	t.proposalTaskLock.RUnlock()
	if !ok {
		return true
	}
	return false
}

func (t *TwoPC) addProposalTask(task *types.ProposalTask) {
	t.proposalTaskLock.Lock()
	_, ok := t.proposalTaskCache[task.TaskId()]
	if !ok {
		t.proposalTaskCache[task.TaskId()] = task
	}
	t.proposalTaskLock.Unlock()
}
func (t *TwoPC) removeProposalTask(taskId string) {
	t.proposalTaskLock.Lock()
	delete(t.proposalTaskCache, taskId)
	t.proposalTaskLock.Unlock()
}



func (t *TwoPC) addTaskResultCh(taskId string, resultCh chan<- *types.TaskConsResult) {
	t.taskResultLock.Lock()
	log.Debugf("AddTaskResultCh taskId: {%s}", taskId)
	t.taskResultChSet[taskId] = resultCh
	t.taskResultLock.Unlock()
}
func (t *TwoPC) removeTaskResultCh(taskId string) {
	t.taskResultLock.Lock()
	log.Debugf("RemoveTaskResultCh taskId: {%s}", taskId)
	delete(t.taskResultChSet, taskId)
	t.taskResultLock.Unlock()
}
func (t *TwoPC) replyTaskConsensusResult(result *types.TaskConsResult) {
	t.taskResultBusCh <- result
}
func (t *TwoPC) handleTaskConsensusResult(result *types.TaskConsResult) {
	t.taskResultLock.Lock()
	log.Debugf("Need SendTaskResultCh taskId: {%s}, result: {%s}", result.TaskId, result.String())
	if ch, ok := t.taskResultChSet[result.TaskId]; ok {
		log.Debugf("Start SendTaskResultCh taskId: {%s}, result: {%s}", result.TaskId, result.String())
		ch <- result
		close(ch)
		delete(t.taskResultChSet, result.TaskId)
	}
	t.taskResultLock.Unlock()
}

func (t *TwoPC) sendReplaySchedTaskToScheduler(replaySchedTask *types.ReplayScheduleTaskWrap) {
	t.replayTaskCh <- replaySchedTask
}

func (t *TwoPC) storeProposalState(proposalState *ctypes.ProposalState) {
	t.state.AddProposalState(proposalState)
}
func (t *TwoPC) removeProposalState(proposalId common.Hash) {
	t.state.CleanProposalState(proposalId)
}
func (t *TwoPC) delProposalStateAndTask(proposalId common.Hash) {
	if state := t.state.GetProposalState(proposalId); t.state.EmptyInfo() != state {
		log.Infof("Start remove proposalState and task cache on Consensus, proposalId {%s}, taskId {%s}", proposalId, state.TaskId)
		t.removeProposalTask(state.GetTaskId())
		t.state.CleanProposalState(proposalId)

	}
}

func (t *TwoPC) sendTaskToTaskManagerForExecute(task *types.DoneScheduleTaskChWrap) {
	t.doneScheduleTaskCh <- task
}

func (t *TwoPC) makeConfirmTaskPeerDesc(proposalId common.Hash) *pb.ConfirmTaskPeerInfo {

	dataSuppliers, powerSuppliers, receivers := make([]*pb.TaskPeerInfo, 0), make([]*pb.TaskPeerInfo, 0), make([]*pb.TaskPeerInfo, 0)

	for _, vote := range t.state.GetPrepareVoteArr(proposalId) {
		if vote.TaskRole == types.DataSupplier && nil != vote.PeerInfo {
			dataSuppliers = append(dataSuppliers, types.ConvertTaskPeerInfo(vote.PeerInfo))
		}
		if vote.TaskRole == types.PowerSupplier && nil != vote.PeerInfo {
			powerSuppliers = append(powerSuppliers, types.ConvertTaskPeerInfo(vote.PeerInfo))
		}
		if vote.TaskRole == types.ResultSupplier && nil != vote.PeerInfo {
			receivers = append(receivers, types.ConvertTaskPeerInfo(vote.PeerInfo))
		}
	}
	owner := t.state.GetSelfPeerInfo(proposalId)
	if nil == owner {
		return nil
	}
	return &pb.ConfirmTaskPeerInfo{
		OwnerPeerInfo:              types.ConvertTaskPeerInfo(owner),
		DataSupplierPeerInfoList:   dataSuppliers,
		PowerSupplierPeerInfoList:  powerSuppliers,
		ResultReceiverPeerInfoList: receivers,
	}
}

func (t *TwoPC) refreshProposalState() {

	for id, proposalState := range t.state.GetProposalStates() {

		if proposalState.IsDeadline() {
			log.Debugf("Started refresh proposalState loop, the proposalState direct be deadline, proposalId: {%s}, taskId: {%s}",
				id.String(), proposalState.TaskId)
			t.handleInvalidProposal(proposalState)
			continue
		}

		switch proposalState.CurrPeriodNum() {
		case ctypes.PeriodPrepare:
			if proposalState.IsPrepareTimeout() {
				log.Debugf("Started refresh proposalState loop, the proposalState was prepareTimeout, change to confirm epoch, proposalId: {%s}, taskId: {%s}",
					id.String(), proposalState.TaskId)
				proposalState.ChangeToConfirm(proposalState.PeriodStartTime + uint64(ctypes.PrepareMsgVotingTimeout.Milliseconds()))
				t.state.UpdateProposalState(proposalState)
			}
		case ctypes.PeriodConfirm:
			if proposalState.IsConfirmTimeout() {
				log.Debugf("Started refresh proposalState loop, the proposalState was confirmTimeout, change to commit epoch, proposalId: {%s}, taskId: {%s}",
					id.String(), proposalState.TaskId)
				proposalState.ChangeToCommit(proposalState.PeriodStartTime + uint64(ctypes.ConfirmMsgVotingTimeout.Milliseconds()))
				t.state.UpdateProposalState(proposalState)
			}
		case ctypes.PeriodCommit:
			if proposalState.IsCommitTimeout() {
				log.Debugf("Started refresh proposalState loop, the proposalState was commitTimeout, change to finished epoch, proposalId: {%s}, taskId: {%s}",
					id.String(), proposalState.TaskId)
				proposalState.ChangeToFinished(proposalState.PeriodStartTime + uint64(ctypes.CommitMsgEndingTimeout.Milliseconds()))
				//t.state.UpdateProposalState(proposalState)
				t.handleInvalidProposal(proposalState)
			}
		case ctypes.PeriodFinished:
			//
			if proposalState.IsDeadline() {
				log.Debugf("Started refresh proposalState loop, the proposalState was finished but coming deadline now, proposalId: {%s}, taskId: {%s}",
					id.String(), proposalState.TaskId)
				t.handleInvalidProposal(proposalState)
			}

		default:
			log.Errorf("Unknown the proposalState period, proposalId: {%s}, taskId: {%s}", id.String(), proposalState.TaskId)
			t.handleInvalidProposal(proposalState)
		}
	}
}

func (t *TwoPC) handleInvalidProposal(proposalState *ctypes.ProposalState) {

	log.Debugf("Call handleInvalidProposal(), handle and clean proposalState and task, proposalId: {%s}, taskId: {%s}, taskDir: {%s}", proposalState.ProposalId, proposalState.TaskId, proposalState.TaskDir.String())

	has, err := t.dataCenter.HasLocalTaskExecute(proposalState.TaskId)
	if nil != err {
		log.Errorf("Failed to query local task exec status with task on handleInvalidProposal(), taskId: {%s}, err: {%s}", proposalState.TaskId, err)
		// 最终 clean some data
		t.delProposalStateAndTask(proposalState.ProposalId)
		return
	}

	if has {
		log.Debugf("The local task have been executing, direct clean proposalStateAndTaskCache of consensus, taskId: {%s}", proposalState.TaskId)
		// 最终 clean some data
		t.delProposalStateAndTask(proposalState.ProposalId)
		return
	}


	if proposalState.TaskDir == types.SendTaskDir {
		// Send consensus result to Scheduler
		t.replyTaskConsensusResult(&types.ConsensusResult{
			TaskConsResult: &types.TaskConsResult{
				TaskId: proposalState.TaskId,
				Status: types.TaskConsensusInterrupt,
				Done:   false,
				Err:    fmt.Errorf("the task proposalState coming deadline"),
			},
		})

	} else {

		task, ok := t.GetRecvTaskWithOk(proposalState.TaskId)
		if !ok {
			log.Errorf("Failed to query recvTaskInfo on consensus.handleInvalidProposal(), taskId: {%s}", proposalState.TaskId)
			return
		}
		eventList, err := t.dataCenter.GetTaskEventList(proposalState.TaskId)
		if nil != err {
			log.Errorf("Failed to GetTaskEventList() on consensus.handleInvalidProposal(), taskId: {%s}, err: {%s}", proposalState.TaskId, err)
			eventList = make([]*libTypes.TaskEvent, 0)
		}
		eventList = append(eventList, &libTypes.TaskEvent{
			Type: evengine.TaskProposalStateDeadline.Type,
			IdentityId: proposalState.TaskOrg.IdentityId,
			TaskId: proposalState.TaskId,
			Content: fmt.Sprintf("%s for myself", evengine.TaskProposalStateDeadline.Msg),
			CreateAt: uint64(timeutils.UnixMsec()),
		})
		taskResultWrap := &types.TaskResultMsgWrap{
			TaskResultMsg: &pb.TaskResultMsg{
				ProposalId: proposalState.ProposalId.Bytes(),
				TaskRole: proposalState.TaskRole.Bytes(),
				TaskId: []byte(proposalState.TaskId),
				Owner: &pb.TaskOrganizationIdentityInfo{
					PartyId: []byte(proposalState.TaskOrg.PartyId),
					Name: []byte(proposalState.TaskOrg.NodeName),
					NodeId: []byte(proposalState.TaskOrg.NodeId),
					IdentityId: []byte(proposalState.TaskOrg.IdentityId),
				},
				TaskEventList: types.ConvertTaskEventArr(eventList),
				CreateAt: uint64(timeutils.UnixMsec()),
				Sign: nil,
			},
		}

		// Send taskResultMsg to task Owner
		pid, err := p2p.HexPeerID(task.TaskData().NodeId)
		if nil == err {
			if err := t.sendTaskResultMsg(pid, taskResultWrap); nil != err {
				log.Error(err)
			}
		}

		t.resourceMng.ReleaseLocalResourceWithTask("on consensus.handleInvalidProposal()", proposalState.TaskId, resource.SetAllReleaseResourceOption())

	}

	// 最终 clean some data
	t.delProposalStateAndTask(proposalState.ProposalId)
}

func (t *TwoPC) storeTaskEvent(pid peer.ID, taskId string, events []*libTypes.TaskEvent) error {
	for _, event := range events {
		if err := t.dataCenter.StoreTaskEvent(event); nil != err {
			log.Error("Failed to store local task event from remote peer", "remote peerId", pid, "taskId", taskId)
		}
	}
	return nil
}

func (t *TwoPC) driveTask(
	pid peer.ID,
	proposalId common.Hash,
	taskDir types.ProposalTaskDir,
	taskState apipb.TaskState,
	taskRole  apipb.TaskRole,
	selfIdentity *apipb.TaskOrganization,
	task *types.Task,
	) {

	log.Debugf("Start to call `driveTask`, proposalId: {%s}, taskId: {%s}, taskDir: {%s}, taskState: {%s}, taskRole: {%s}, myselfIdentityId: {%s}",
		proposalId.String(), task.TaskId(), taskDir.String(), taskState.String(), taskRole.String(), selfIdentity.IdentityId)

	selfVotePeerInfo := t.state.GetSelfPeerInfo(proposalId)
	if nil == selfVotePeerInfo {
		log.Errorf("Failed to find local cache about prepareVote myself internal resource, proposalId: {%s}, taskId: {%s}, taskDir: {%s}, taskState: {%s}, taskRole: {%s}, myselfIdentityId: {%s}",
			proposalId.String(), task.TaskId(), taskDir.String(), taskState.String(), taskRole.String(), selfIdentity.IdentityId)
		return
	}

	confirmTaskPeerInfo := t.state.GetConfirmTaskPeerInfo(proposalId)
	if nil == confirmTaskPeerInfo {
		log.Errorf("Failed to find local cache about prepareVote all peer resource {externalIP:externalPORT}, proposalId: {%s}, taskId: {%s}, taskDir: {%s}, taskState: {%s}, taskRole: {%s}, myselfIdentityId: {%s}",
			proposalId.String(), task.TaskId(), taskDir.String(), taskState.String(), taskRole.String(), selfIdentity.IdentityId)
		return
	}

	// Store task exec status
	if err := t.dataCenter.StoreLocalTaskExecuteStatus(task.TaskId()); nil != err {
		log.Errorf("Failed to store local task about exec status, proposalId: {%s}, taskId: {%s}, taskDir: {%s}, taskState: {%s}, taskRole: {%s}, myselfIdentityId: {%s}, err: {%s}",
			proposalId.String(), task.TaskId(), taskDir.String(), taskState.String(), taskRole.String(), selfIdentity.IdentityId, err)
		return
	}

	// Send task to TaskManager to execute
	taskWrap := &types.DoneScheduleTaskChWrap{
		ProposalId: proposalId,
		SelfTaskRole: taskRole,
		SelfIdentity: selfIdentity,
		Task: &types.ConsensusScheduleTask{
			TaskDir:   taskDir,
			TaskState: taskState,
			SchedTask: task,
			SelfVotePeerInfo: &types.PrepareVoteResource{
				Id: selfVotePeerInfo.Id,
				Ip: selfVotePeerInfo.Ip,
				Port: selfVotePeerInfo.Port,
				PartyId: selfVotePeerInfo.PartyId,
			},
			Resources: confirmTaskPeerInfo,
		},
		ResultCh: make(chan *types.TaskResultMsgWrap, 0),
	}
	// 发给 taskManager 去执行 task
	t.sendTaskToTaskManagerForExecute(taskWrap)
	go func() {
		if taskDir == types.RecvTaskDir {
			if taskResultWrap, ok := <-taskWrap.ResultCh; ok {
				if err := t.sendTaskResultMsg(pid, taskResultWrap); nil != err {
					log.Error(err)
				}
				t.resourceMng.ReleaseLocalResourceWithTask("on consensus.driveTask()", task.TaskId(), resource.SetAllReleaseResourceOption())
				//// clean some data
				//t.delProposalStateAndTask(proposalId)
			}
		} else {
			<-taskWrap.ResultCh  // publish taskInfo to dataCenter done ..
			//// clean local proposalState and task cache
			//t.delProposalStateAndTask(proposalId)
		}
	}()
}

func (t *TwoPC) sendPrepareMsg(proposalId common.Hash, task *types.Task, startTime uint64) error {

	sender := task.TaskSender()

	sendTaskFn := func(wg *sync.WaitGroup, sender, receiver *apipb.TaskOrganization, senderRole, receiverRole apipb.TaskRole, errCh chan<- error) {

		defer wg.Done()

		var pid, err = p2p.HexPeerID(receiver.NodeId)
		if nil != err {
			errCh <- fmt.Errorf("failed to nodeId => peerId, proposalId: %s, taskId: %s, other peer taskRole: %s, other peer taskPartyId: %s, identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			return
		}

		prepareMsg, err := makePrepareMsg(proposalId, senderRole, receiverRole, sender.GetPartyId(), receiver.GetPartyId(), task, startTime)

		if nil != err {
			errCh <- fmt.Errorf("failed to make prepareMsg, proposalId: %s, taskId: %s, other peer taskRole: %s, other peer taskPartyId: %s, identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			return
		}

		var sendErr error
		if types.IsSameTaskOrg(sender, receiver) {
			sendErr = t.sendLocalPrepareMsg(pid, prepareMsg)
		} else {
			sendErr = handler.SendTwoPcPrepareMsg(context.TODO(), t.p2p, pid, prepareMsg)
		}

		if nil != sendErr {
			errCh <- fmt.Errorf("failed to call `SendTwoPcPrepareMsg` proposalId: %s, taskId: %s, other peer taskRole: %s, other peer taskPartyId: %s, identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, sendErr)
			return
		}

		log.Debugf("Succceed to call `SendTwoPcPrepareMsg` proposalId: %s, taskId: %s, other peer taskRole: %s, other peer taskPartyId: %s, identityId: %s, pid: %s",
			proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid)
	}

	size := (len(task.TaskData().GetDataSuppliers())) + len(task.TaskData().GetPowerSuppliers()) + len(task.TaskData().GetReceivers())
	errCh := make(chan error, size)
	var wg sync.WaitGroup

	for i := 0; i < len(task.TaskData().GetDataSuppliers()); i++ {

		wg.Add(1)
		dataSupplier := task.TaskData().GetDataSuppliers()[i]
		receiver := dataSupplier.GetMemberInfo()
		go sendTaskFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_DataSupplier, errCh)

	}
	for i := 0; i < len(task.TaskData().GetPowerSuppliers()); i++ {

		wg.Add(1)
		powerSupplier := task.TaskData().GetPowerSuppliers()[i]
		receiver := powerSupplier.GetOrganization()
		go sendTaskFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_PowerSupplier, errCh)

	}

	for i := 0; i < len(task.TaskData().GetReceivers()); i++ {

		wg.Add(1)
		receiver := task.TaskData().GetReceivers()[i]
		go sendTaskFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_Receiver, errCh)
	}

	wg.Wait()
	close(errCh)

	errStrs := make([]string, 0)

	for err := range errCh {
		if nil != err {
			errStrs = append(errStrs, err.Error())
		}
	}
	if len(errStrs) != 0 {
		return fmt.Errorf(
			"\n######################################################## \n%s\n########################################################\n",
			strings.Join(errStrs, "\n"))
	}
	return nil
}

func (t *TwoPC) sendConfirmMsg(proposalId common.Hash, task *types.Task, startTime uint64) error {

	peers:= t.makeConfirmTaskPeerDesc(proposalId)
	sender := task.TaskSender()

	sendConfirmMsgFn := func(wg *sync.WaitGroup, sender, receiver *apipb.TaskOrganization, senderRole, receiverRole apipb.TaskRole, errCh chan<- error) {

		defer wg.Done()

		pid, err := p2p.HexPeerID(receiver.NodeId)
		if nil != err {
			errCh <- fmt.Errorf("failed to nodeId => peerId, proposalId: %s, taskId: %s, other peer's taskRole: %s, other peer's partyId: %s, other identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			return
		}

		confirmMsg := makeConfirmMsg(proposalId, senderRole, receiverRole, sender.GetPartyId(), receiver.GetPartyId(), task, peers, startTime)

		var sendErr error
		if types.IsSameTaskOrg(sender, receiver) {
			sendErr = t.sendLocalConfirmMsg(pid, confirmMsg)
		} else {
			sendErr = handler.SendTwoPcConfirmMsg(context.TODO(), t.p2p, pid, confirmMsg)
		}

		// Send the ConfirmMsg to other peer
		if nil != sendErr {
			errCh <- fmt.Errorf("failed to call`SendTwoPcConfirmMsg` proposalId: %s, taskId: %s,other peer's taskRole: %s, other peer's partyId: %s, other identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			errCh <- err
			return
		}

		log.Debugf("Succceed to call`SendTwoPcConfirmMsg` proposalId: %s, taskId: %s,other peer's taskRole: %s, other peer's partyId: %s, other identityId: %s, pid: %s",
			proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid)

	}

	size := (len(task.TaskData().GetDataSuppliers())) + len(task.TaskData().GetPowerSuppliers()) + len(task.TaskData().GetReceivers())
	errCh := make(chan error, size)
	var wg sync.WaitGroup

	for i := 0; i < len(task.TaskData().GetDataSuppliers()); i++ {

		wg.Add(1)
		dataSupplier := task.TaskData().GetDataSuppliers()[i]
		receiver := dataSupplier.GetMemberInfo()
		go sendConfirmMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_DataSupplier, errCh)

	}
	for i := 0; i < len(task.TaskData().GetPowerSuppliers()); i++ {

		wg.Add(1)
		powerSupplier := task.TaskData().GetPowerSuppliers()[i]
		receiver := powerSupplier.GetOrganization()
		go sendConfirmMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_PowerSupplier, errCh)

	}

	for i := 0; i < len(task.TaskData().GetReceivers()); i++ {

		wg.Add(1)
		receiver := task.TaskData().GetReceivers()[i]
		go sendConfirmMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_Receiver, errCh)
	}

	wg.Wait()
	close(errCh)

	errStrs := make([]string, 0)

	for err := range errCh {
		if nil != err {
			errStrs = append(errStrs, err.Error())
		}
	}
	if len(errStrs) != 0 {
		return fmt.Errorf(
			"\n######################################################## \n%s\n########################################################\n",
			strings.Join(errStrs, "\n"))
	}
	return nil
}

func (t *TwoPC) sendCommitMsg(proposalId common.Hash, task *types.Task, startTime uint64) error {

	sender := task.TaskSender()

	sendCommitMsgFn := func(wg *sync.WaitGroup, sender, receiver *apipb.TaskOrganization, senderRole, receiverRole apipb.TaskRole, errCh chan<- error) {

		defer wg.Done()

		pid, err := p2p.HexPeerID(receiver.NodeId)
		if nil != err {
			errCh <- fmt.Errorf("failed to nodeId => peerId, proposalId: %s, taskId: %s, other peer's taskRole: %s, other peer's partyId: %s, identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			return
		}

		commitMsg := makeCommitMsg(proposalId, senderRole, receiverRole, sender.GetPartyId(), receiver.GetPartyId(), task, startTime)


		var sendErr error
		if types.IsSameTaskOrg(sender, receiver) {
			sendErr = t.sendLocalCommitMsg(pid, commitMsg)
		} else {
			sendErr = handler.SendTwoPcCommitMsg(context.TODO(), t.p2p, pid, commitMsg)
		}

		// Send the ConfirmMsg to other peer
		if nil != sendErr {
			errCh <- fmt.Errorf("failed to call`SendTwoPcCommitMsg` proposalId: %s, taskId: %s,  other peer's taskRole: %s, other peer's partyId: %s, identityId: %s, pid: %s, err: %s",
				proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid, err)
			errCh <- err
			return
		}

		log.Debugf("Succceed to call`SendTwoPcCommitMsg` proposalId: %s, taskId: %s,  other peer's taskRole: %s, other peer's partyId: %s, identityId: %s, pid: %s",
			proposalId.String(), task.TaskId(), receiverRole.String(), receiver.GetPartyId(), receiver.GetIdentityId(), pid)

	}


	size := (len(task.TaskData().GetDataSuppliers())) + len(task.TaskData().GetPowerSuppliers()) + len(task.TaskData().GetReceivers())
	errCh := make(chan error, size)
	var wg sync.WaitGroup

	for i := 0; i < len(task.TaskData().GetDataSuppliers()); i++ {

		wg.Add(1)
		dataSupplier := task.TaskData().GetDataSuppliers()[i]
		receiver := dataSupplier.GetMemberInfo()
		go sendCommitMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_DataSupplier, errCh)

	}
	for i := 0; i < len(task.TaskData().GetPowerSuppliers()); i++ {

		wg.Add(1)
		powerSupplier := task.TaskData().GetPowerSuppliers()[i]
		receiver := powerSupplier.GetOrganization()
		go sendCommitMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_PowerSupplier, errCh)

	}

	for i := 0; i < len(task.TaskData().GetReceivers()); i++ {

		wg.Add(1)
		receiver := task.TaskData().GetReceivers()[i]
		go sendCommitMsgFn(&wg,sender, receiver, apipb.TaskRole_TaskRole_Sender, apipb.TaskRole_TaskRole_Receiver, errCh)
	}

	wg.Wait()
	close(errCh)

	errStrs := make([]string, 0)

	for err := range errCh {
		if nil != err {
			errStrs = append(errStrs, err.Error())
		}
	}
	if len(errStrs) != 0 {
		return fmt.Errorf(
			"\n######################################################## \n%s\n########################################################\n",
			strings.Join(errStrs, "\n"))
	}
	return nil
}

