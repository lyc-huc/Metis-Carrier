package rawdb

import (
	"bytes"
	"fmt"
	"github.com/RosettaFlow/Carrier-Go/common/bytesutil"
	apicommonpb "github.com/RosettaFlow/Carrier-Go/lib/common"
	libtypes "github.com/RosettaFlow/Carrier-Go/lib/types"
	"github.com/RosettaFlow/Carrier-Go/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gogo/protobuf/proto"
	leveldberr "github.com/syndtr/goleveldb/leveldb/errors"
	"strings"
)

var (
	ErrNotFound = leveldberr.ErrNotFound
)

func IsNoDBNotFoundErr(err error) bool {
	return nil != err && err != ErrNotFound
}
func IsDBNotFoundErr(err error) bool {
	return nil != err && err == ErrNotFound
}

// Resources that operate the organization's jobNode services
func StoreNodeResource(db KeyValueStore, resource *types.LocalResourceTable) error {

	item_key := GetNodeResourceKey(resource.GetNodeId())
	val, err := rlp.EncodeToBytes(resource)
	if nil != err {
		return err
	}
	return db.Put(item_key, val)
}

func StoreNodeResources(db KeyValueStore, resources []*types.LocalResourceTable) error {

	for _, resource := range resources {
		key := GetNodeResourceKey(resource.GetNodeId())
		val, err := rlp.EncodeToBytes(resource)
		if nil != err {
			return err
		}

		if err := db.Put(key, val); nil != err {
			return err
		}
	}
	return nil
}

func RemoveNodeResource(db KeyValueStore, resourceId string) error {
	key := GetNodeResourceKey(resourceId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryNodeResource(db DatabaseReader, resourceId string) (*types.LocalResourceTable, error) {
	key := GetNodeResourceKey(resourceId)
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}

	var resource types.LocalResourceTable
	if err := rlp.DecodeBytes(vb, &resource); nil != err {
		return nil, err
	}
	return &resource, nil
}

func QueryNodeResources(db KeyValueStore) ([]*types.LocalResourceTable, error) {

	prefix := GetNodeResourceKeyPrefix()
	it := db.NewIteratorWithPrefixAndStart(prefix, nil)
	defer it.Release()

	arr := make([]*types.LocalResourceTable, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// prefix + jobNodeId -> LocalResourceTable
			var resource types.LocalResourceTable
			if err := rlp.DecodeBytes(it.Value(), &resource); nil != err {
				return nil, err
			}
			arr = append(arr, &resource)
		}
	}

	if len(arr) == 0 {
		return nil, ErrNotFound
	}

	return arr, nil
}

// Operation resource slot unit definition
func StoreNodeResourceSlotUnit(db DatabaseWriter, slot *types.Slot) error {
	val, err := rlp.EncodeToBytes(slot)
	if nil != err {
		return err
	}
	return db.Put(GetNodeResourceSlotUnitKey(), val)
}

func RemoveNodeResourceSlotUnit(db KeyValueStore) error {
	key := GetNodeResourceSlotUnitKey()
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryNodeResourceSlotUnit(db DatabaseReader) (*types.Slot, error) {
	has, err := db.Has(GetNodeResourceSlotUnitKey())
	if IsNoDBNotFoundErr(err) {
		return nil, err
	}
	if !has {
		return nil, ErrNotFound
	}
	b, err := db.Get(GetNodeResourceSlotUnitKey())
	if nil != err {
		return nil, err
	}
	var slot *types.Slot
	if err := rlp.DecodeBytes(b, &slot); nil != err {
		return nil, err
	}
	return slot, nil
}

// Operate the information of the jobNode resources being used by the local task
func StoreLocalTaskPowerUsed(db KeyValueStore, taskPowerUsed *types.LocalTaskPowerUsed) error {
	// prefix + taskId + partyId -> LocalTaskPowerUsed
	key := GetLocalTaskPowerUsedKey(taskPowerUsed.GetTaskId(), taskPowerUsed.GetPartyId())
	val, err := rlp.EncodeToBytes(taskPowerUsed)
	if nil != err {
		return err
	}
	log.Debugf("Call StoreLocalTaskPowerUsed, taskId: {%s}, partyId: {%s}, used: {%s}", taskPowerUsed.GetTaskId(), taskPowerUsed.GetPartyId(), taskPowerUsed.String())
	return db.Put(key, val)
}

func StoreLocalTaskPowerUseds(db KeyValueStore, taskPowerUseds []*types.LocalTaskPowerUsed) error {
	for _, used := range taskPowerUseds {
		key := GetLocalTaskPowerUsedKey(used.GetTaskId(), used.GetPartyId())
		val, err := rlp.EncodeToBytes(used)
		if nil != err {
			return err
		}

		if err := db.Put(key, val); nil != err {
			return err
		}
	}
	return nil
}

func HasLocalTaskPowerUsed(db DatabaseReader, taskId, partyId string) (bool, error) {

	has, err := db.Has(GetLocalTaskPowerUsedKey(taskId, partyId))

	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err), !has:
		return false, nil
	}
	return true, nil
}

func RemoveLocalTaskPowerUsed(db KeyValueStore, taskId, partyId string) error {
	key := GetLocalTaskPowerUsedKey(taskId, partyId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func RemoveLocalTaskPowerUsedByTaskId(db KeyValueStore, taskId string) error {
	it := db.NewIteratorWithPrefixAndStart(GetLocalTaskPowerUsedKeyPrefixByTaskId(taskId), nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			db.Delete(key)
		}
	}

	return nil
}

func QueryLocalTaskPowerUsed(db DatabaseReader, taskId, partyId string) (*types.LocalTaskPowerUsed, error) {
	// prefix + taskId + partyId -> LocalTaskPowerUsed
	key := GetLocalTaskPowerUsedKey(taskId, partyId)
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}

	var taskPowerUsed types.LocalTaskPowerUsed

	if err := rlp.DecodeBytes(vb, &taskPowerUsed); nil != err {
		return nil, err
	}
	used := &taskPowerUsed
	log.Debugf("Call QueryLocalTaskPowerUsed, taskId: {%s}, partyId: {%s}, used: {%s}", taskId, partyId, used.String())
	return &taskPowerUsed, nil
}

func QueryLocalTaskPowerUsedsByTaskId(db KeyValueStore, taskId string) ([]*types.LocalTaskPowerUsed, error) {
	// prefix + taskId + partyId -> LocalTaskPowerUsed
	it := db.NewIteratorWithPrefixAndStart(GetLocalTaskPowerUsedKeyPrefixByTaskId(taskId), nil)
	defer it.Release()

	arr := make([]*types.LocalTaskPowerUsed, 0)
	for it.Next() {
		if value := it.Value(); len(value) != 0 {
			var taskPowerUsed types.LocalTaskPowerUsed
			if err := rlp.DecodeBytes(value, &taskPowerUsed); nil != err {
				return nil, err
			}
			arr = append(arr, &taskPowerUsed)
		}
	}

	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func QueryLocalTaskPowerUseds(db KeyValueStore) ([]*types.LocalTaskPowerUsed, error) {
	// prefix + taskId + partyId -> LocalTaskPowerUsed
	it := db.NewIteratorWithPrefixAndStart(GetLocalTaskPowerUsedKeyPrefix(), nil)
	defer it.Release()

	arr := make([]*types.LocalTaskPowerUsed, 0)
	for it.Next() {
		if value := it.Value(); len(value) != 0 {
			var taskPowerUsed types.LocalTaskPowerUsed
			if err := rlp.DecodeBytes(value, &taskPowerUsed); nil != err {
				return nil, err
			}
			arr = append(arr, &taskPowerUsed)
		}
	}

	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func StoreJobNodeTaskPartyId(db KeyValueStore, jobNodeId, taskId, partyId string) error {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return err
	}
	var partyIdArr []string
	if !has {
		partyIdArr = []string{partyId}
	} else {

		val, err := db.Get(key)
		if nil != err {
			return err
		}
		if err := rlp.DecodeBytes(val, &partyIdArr); nil != err {
			return err
		}

		var find bool
		for _, id := range partyIdArr {
			if id == partyId {
				find = true
				break
			}
		}
		if !find {
			partyIdArr = append(partyIdArr, partyId)
		}
	}
	val, err := rlp.EncodeToBytes(partyIdArr)
	if nil != err {
		return err
	}
	log.Debugf("Call StoreJobNodeTaskPartyId, jobNodeId: {%s}, taskId: {%s}, partyId: {%s}, partyIds: %s", jobNodeId, taskId, partyId, partyIdArr)
	return db.Put(key, val)
}

func RemoveJobNodeTaskPartyId(db KeyValueStore, jobNodeId, taskId, partyId string) error {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	val, err := db.Get(key)

	var partyIdArr []string
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err):
		return nil
	case nil == err && len(val) != 0:
		if err := rlp.DecodeBytes(val, &partyIdArr); nil != err {
			return err
		}
	}

	//for i := 0; i < len(partyIdArr); i++ {
	//
	//	id := partyIdArr[i]
	//	if id == partyId {
	//		partyIdArr = append(partyIdArr[:i], partyIdArr[i+1:]...)
	//		i--
	//	}
	//}

	for i, id := range partyIdArr {
		if id == partyId {
			partyIdArr = append(partyIdArr[:i], partyIdArr[i+1:]...)
			break
		}
	}

	if len(partyIdArr) == 0 {
		log.Debugf("Call RemoveJobNodeTaskPartyId [clean all partyIds], jobNodeId: {%s}, taskId: {%s}, partyId: {%s}", jobNodeId, taskId, partyId)
		return db.Delete(key)
	}
	val, err = rlp.EncodeToBytes(partyIdArr)
	if nil != err {
		return err
	}
	log.Debugf("Call RemoveJobNodeTaskPartyId, jobNodeId: {%s}, taskId: {%s}, partyId: {%s}, partyIds: %s", jobNodeId, taskId, partyId, partyIdArr)
	return db.Put(key, val)
}

func RemoveJobNodeTaskIdAllPartyIds(db KeyValueStore, jobNodeId, taskId string) error {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryJobNodeRunningTaskIds(db KeyValueStore, jobNodeId string) ([]string, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	prefixAndJobNodeId := GetJobNodeTaskPartyIdsKeyPrefixByJobNodeId(jobNodeId)
	it := db.NewIteratorWithPrefixAndStart(prefixAndJobNodeId, nil)
	defer it.Release()

	arr := make([]string, 0)
	tmp := make(map[string]struct{}, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// key len == len(prefix) + len([]byte(jobNodeId)) + len([]byte(taskId))
			taskId := string(it.Key()[len(prefixAndJobNodeId):])
			if _, ok := tmp[taskId]; !ok {
				tmp[taskId] = struct{}{}
				arr = append(arr, taskId)
			}
		}
	}

	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	log.Debugf("Call QueryJobNodeRunningTaskIds, jobNodeId: {%s}, taskIds: %s", jobNodeId, strings.Join(arr, ","))
	return arr, nil
}

func QueryJobNodeRunningTaskIdCount(db KeyValueStore, jobNodeId string) (uint32, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	prefixAndJobNodeId := GetJobNodeTaskPartyIdsKeyPrefixByJobNodeId(jobNodeId)
	it := db.NewIteratorWithPrefixAndStart(prefixAndJobNodeId, nil)
	defer it.Release()

	var count uint32
	tmp := make(map[string]struct{}, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// key len == len(prefix) + len([]byte(jobNodeId)) + len([]byte(taskId))
			taskId := string(it.Key()[len(prefixAndJobNodeId):])
			if _, ok := tmp[taskId]; !ok {
				tmp[taskId] = struct{}{}
				count ++
			}
		}
	}
	log.Debugf("Call QueryJobNodeRunningTaskCount, jobNodeId: {%s}, taskIds count: %d", jobNodeId, count)
	return count, nil
}

func QueryJobNodeTaskAllPartyIds(db KeyValueStore, jobNodeId, taskId string) ([]string, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	val, err := db.Get(key)
	var partyIdArr []string
	switch {
	case IsNoDBNotFoundErr(err):
		return nil, err
	case IsDBNotFoundErr(err):
		return nil, ErrNotFound
	case nil == err && len(val) != 0:
		if err := rlp.DecodeBytes(val, &partyIdArr); nil != err {
			return nil, err
		}
	}
	return partyIdArr, nil
}

func HasJobNodeRunningTaskId(db DatabaseReader, jobNodeId, taskId string) (bool, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err):
		return false, nil
	case nil == err && !has:
		return false, nil
	}
	return true, nil
}

func HasJobNodeTaskPartyId (db DatabaseReader, jobNodeId, taskId, partyId string) (bool, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	val, err := db.Get(key)
	var partyIdArr []string
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err):
		return false, nil
	case nil == err && len(val) != 0:
		if err := rlp.DecodeBytes(val, &partyIdArr); nil != err {
			return false, err
		}
	}

	for _, id := range partyIdArr {
		if id == partyId {
			return true, nil
		}
	}
	return false, nil
}

func QueryJobNodeTaskPartyIdCount(db DatabaseReader, jobNodeId, taskId string) (uint32, error) {
	// prefix + jobNodeId + taskId -> [partyId, ..., partyId]
	key := GetJobNodeTaskPartyIdsKey(jobNodeId, taskId)
	val, err := db.Get(key)

	var partyIdArr []string
	switch {
	case IsNoDBNotFoundErr(err):
		return 0, err
	case IsDBNotFoundErr(err):
		return 0, nil
	case nil == err && len(val) != 0:
		if err := rlp.DecodeBytes(val, &partyIdArr); nil != err {
			return 0, err
		}
	}
	return uint32(len(partyIdArr)), nil
}

// about jobNode history task
func StoreJobNodeHistoryTaskId (db KeyValueStore, jobNodeId, taskId string) error {

	// prefix + jobNodeId + taskId -> index
	item_key := GetJobNodeHistoryTaskKey(jobNodeId, taskId)
	has, err := db.Has(item_key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case nil == err && has:
		return nil   // It have been exists, don't inscrease count
	}

	// When taskId have not on jobNode, inscrease jobNode taskId count
	// and put taskId on jobNodeId mapping.
	//
	// prefix + jobNodeId -> history task count
	count_key := GetJobNodeHistoryTaskCountKey(jobNodeId)
	count_val, err := db.Get(count_key)

	var count uint32

	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err):
		// do nothing
	case nil == err && len(count_val) != 0:
		count = bytesutil.BytesToUint32(count_val)
	}
	count ++

	count_val = bytesutil.Uint32ToBytes(count)

	// First: put taskId on jobNode mapping.
	if err := db.Put(item_key, count_val); nil != err {
		return err
	}
	log.Debugf("InscreaseJobNodeHistoryTaskCount, jobNodeId: {%s}, taskId: {%s}, count: {%d}", jobNodeId, taskId, count)
	// Second: inscease taskId count on jobNode.
	return db.Put(count_key, count_val)
}

func HasJobNodeHistoryTaskId (db DatabaseReader,jobNodeId, taskId string) (bool, error) {

	item_key := GetJobNodeHistoryTaskKey(jobNodeId, taskId)
	has, err := db.Has(item_key)
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err):
		return false, nil
	case nil == err && !has:
		return false, nil
	}
	return true, nil
}

func QueryJobNodeHistoryTaskCount (db KeyValueStore, jobNodeId string) (uint32, error) {
	// prefix + jobNodeId -> history task count
	key := GetJobNodeHistoryTaskCountKey(jobNodeId)
	val, err := db.Get(key)

	var count uint32

	switch {
	case IsNoDBNotFoundErr(err):
		return 0, err
	case IsDBNotFoundErr(err):
		return 0, nil
	case nil == err && len(val) != 0:
		count = bytesutil.BytesToUint32(val)
	}
	return count, nil
}

// Operation local dataNode resource information
func StoreDataResourceTable(db KeyValueStore, dataResourceTable *types.DataResourceTable) error {

	key := GetDataResourceTableKey(dataResourceTable.GetNodeId())
	val, err := rlp.EncodeToBytes(dataResourceTable)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func StoreDataResourceTables(db KeyValueStore, dataResourceTables []*types.DataResourceTable) error {

	for _, dataResourceTable := range dataResourceTables {
		key := GetDataResourceTableKey(dataResourceTable.GetNodeId())
		val, err := rlp.EncodeToBytes(dataResourceTable)
		if nil != err {
			return err
		}
		if err := db.Put(key, val); nil != err {
			return err
		}
	}
	return nil
}

func RemoveDataResourceTable(db KeyValueStore, nodeId string) error {
	key := GetDataResourceTableKey(nodeId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryDataResourceTable(db DatabaseReader, nodeId string) (*types.DataResourceTable, error) {
	key := GetDataResourceTableKey(nodeId)
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}

	var dataResourceTable types.DataResourceTable

	if err := rlp.DecodeBytes(vb, &dataResourceTable); nil != err {
		return nil, err
	}
	return &dataResourceTable, nil
}

func QueryDataResourceTables(db KeyValueStore) ([]*types.DataResourceTable, error) {

	prefix := GetDataResourceTableKeyPrefix()
	it := db.NewIteratorWithPrefixAndStart(prefix, nil)
	defer it.Release()

	arr := make([]*types.DataResourceTable, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// prefix + jobNodeId -> LocalResourceTable
			var dataResourceTable types.DataResourceTable
			if err := rlp.DecodeBytes(it.Value(), &dataResourceTable); nil != err {
				return nil, err
			}
			arr = append(arr, &dataResourceTable)
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

// The dataNode service information where the operation original file ID is located (originid - > {nodeid / metadataid / filepath})
func StoreDataResourceFileUpload(db KeyValueStore, dataResourceFileUpload *types.DataResourceFileUpload) error {

	key := GetDataResourceFileUploadKey(dataResourceFileUpload.GetOriginId())
	val, err := rlp.EncodeToBytes(dataResourceFileUpload)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func StoreDataResourceFileUploads(db KeyValueStore, dataResourceDataUseds []*types.DataResourceFileUpload) error {

	for _, dataResourceDataUsed := range dataResourceDataUseds {
		key := GetDataResourceFileUploadKey(dataResourceDataUsed.GetOriginId())
		val, err := rlp.EncodeToBytes(dataResourceDataUsed)
		if nil != err {
			return err
		}

		if err := db.Put(key, val); nil != err {
			return err
		}
	}
	return nil
}

func RemoveDataResourceFileUpload(db KeyValueStore, originId string) error {

	key := GetDataResourceFileUploadKey(originId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryDataResourceFileUpload(db DatabaseReader, originId string) (*types.DataResourceFileUpload, error) {
	key := GetDataResourceFileUploadKey(originId)
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}

	var dataResourceDataUsed types.DataResourceFileUpload

	if err := rlp.DecodeBytes(vb, &dataResourceDataUsed); nil != err {
		return nil, err
	}
	return &dataResourceDataUsed, nil
}

func QueryDataResourceFileUploads(db KeyValueStore) ([]*types.DataResourceFileUpload, error) {

	prefix := GetDataResourceFileUploadKeyPrefix()
	it := db.NewIteratorWithPrefixAndStart(prefix, nil)
	defer it.Release()

	arr := make([]*types.DataResourceFileUpload, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// prefix + originId -> DataResourceFileUpload{originId, dataNodeId, metaDataId, filePath}
			var dataResourceDataUsed types.DataResourceFileUpload
			if err := rlp.DecodeBytes(it.Value(), &dataResourceDataUsed); nil != err {
				return nil, err
			}
			arr = append(arr, &dataResourceDataUsed)
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func StoreJobNodeIdByPowerId(db DatabaseWriter, powerId, jobNodeId string) error {
	key := GetPowerIdJobNodeIdMapingKey(powerId)
	index, err := rlp.EncodeToBytes(jobNodeId)
	if nil != err {
		return err
	}
	return db.Put(key, index)
}

func RemoveJobNodeIdByPowerId(db KeyValueStore, powerId string) error {
	key := GetPowerIdJobNodeIdMapingKey(powerId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryJobNodeIdByPowerId(db DatabaseReader, powerId string) (string, error) {
	key := GetPowerIdJobNodeIdMapingKey(powerId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return "", err
	}

	if !has {
		return "", ErrNotFound
	}
	idsByte, err := db.Get(key)
	if nil != err {
		return "", err
	}
	var jobNodeId string
	if err := rlp.DecodeBytes(idsByte, &jobNodeId); nil != err {
		return "", err
	}
	return jobNodeId, nil
}

func StoreDataResourceDiskUsed(db DatabaseWriter, dataResourceDiskUsed *types.DataResourceDiskUsed) error {
	key := GetDataResourceDiskUsedKey(dataResourceDiskUsed.GetMetadataId())
	val, err := rlp.EncodeToBytes(dataResourceDiskUsed)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func RemoveDataResourceDiskUsed(db KeyValueStore, metadataId string) error {
	key := GetDataResourceDiskUsedKey(metadataId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func QueryDataResourceDiskUsed(db DatabaseReader, metadataId string) (*types.DataResourceDiskUsed, error) {
	key := GetDataResourceDiskUsedKey(metadataId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return nil, err
	}

	if !has {
		return nil, ErrNotFound
	}
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}
	var dataResourceDiskUsed types.DataResourceDiskUsed
	if err := rlp.DecodeBytes(vb, &dataResourceDiskUsed); nil != err {
		return nil, err
	}
	return &dataResourceDiskUsed, nil
}

func StoreLocalTaskExecuteStatusValConsByPartyId(db DatabaseWriter, taskId, partyId string) error {
	return db.Put(GetLocalTaskExecuteStatus(taskId, partyId), GetLocalTaskExecuteStatusValCons())
}
func StoreLocalTaskExecuteStatusValExecByPartyId(db DatabaseWriter, taskId, partyId string) error {
	return db.Put(GetLocalTaskExecuteStatus(taskId, partyId), GetLocalTaskExecuteStatusValExec())
}

func RemoveLocalTaskExecuteStatusByPartyId(db KeyValueStore, taskId, partyId string) error {
	key := GetLocalTaskExecuteStatus(taskId, partyId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func HasLocalTaskExecuteStatusParty(db KeyValueStore, taskId string) (bool, error) {
	prefix := append(localTaskExecuteStatusKeyPrefix, []byte(taskId)...)
	it := db.NewIteratorWithPrefixAndStart(prefix, nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			has, err := db.Has(key)
			if IsNoDBNotFoundErr(err) {
				return false, err
			}
			// As long as there is a K-V existence,
			// it is a existence about task party execStatus
			if has {
				return true, nil
			}
		}
	}
	return false, nil
}

func HasLocalTaskExecuteStatusByPartyId(db DatabaseReader, taskId, partyId string) (bool, error) {
	key := GetLocalTaskExecuteStatus(taskId, partyId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return false, err
	}
	if !has {
		return false, nil
	}
	return true, nil
}

func HasLocalTaskExecuteStatusValConsByPartyId(db DatabaseReader, taskId, partyId string) (bool, error) {
	key := GetLocalTaskExecuteStatus(taskId, partyId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return false, err
	}
	if !has {
		return false, nil
	}

	vb, err := db.Get(key)
	if nil != err {
		return false, err
	}
	if bytes.Compare(vb, GetLocalTaskExecuteStatusValCons()) != 0 {
		return false, nil
	}
	return true, nil
}
func HasLocalTaskExecuteStatusValExecByPartyId(db DatabaseReader, taskId, partyId string) (bool, error) {
	key := GetLocalTaskExecuteStatus(taskId, partyId)
	has, err := db.Has(key)
	if IsNoDBNotFoundErr(err) {
		return false, err
	}
	if !has {
		return false, nil
	}

	vb, err := db.Get(key)
	if nil != err {
		return false, err
	}
	if bytes.Compare(vb, GetLocalTaskExecuteStatusValExec()) != 0 {
		return false, nil
	}
	return true, nil
}

func StoreUserMetadataAuthIdByMetadataId(db DatabaseWriter, userType apicommonpb.UserType, user, metadataId, metadataAuthId string) error {

	key := GetUserMetadataAuthByMetadataIdKey(userType, user, metadataId)
	val, err := rlp.EncodeToBytes(metadataAuthId)
	if nil != err {
		return err
	}

	log.Debugf("Store metadataAuth, userType: {%s}, user: {%s}, metadataId: {%s}, metadataAauthId: {%s}", userType.String(), user, metadataId, metadataAuthId)
	return db.Put(key, val)
}

func QueryUserMetadataAuthIdByMetadataId(db DatabaseReader, userType apicommonpb.UserType, user, metadataId string) (string, error) {
	key := GetUserMetadataAuthByMetadataIdKey(userType, user, metadataId)

	val, err := db.Get(key)
	if nil != err {
		return "", err
	}

	var metadataAuthId string
	if err = rlp.DecodeBytes(val, &metadataAuthId); nil != err {
		return "", err
	}

	log.Debugf("Query metadataAuthId, userType: {%s}, user: {%s}, metadataId: {%s}, return metadataAauthId: {%s}", userType.String(), user, metadataId, metadataAuthId)

	if "" == metadataAuthId {
		return "", ErrNotFound
	}
	return metadataAuthId, nil
}

func HasUserMetadataAuthIdByMetadataId(db DatabaseReader, userType apicommonpb.UserType, user, metadataId string) (bool, error) {
	key := GetUserMetadataAuthByMetadataIdKey(userType, user, metadataId)

	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err), nil == err && !has:
		return false, nil
	}
	log.Debugf("Has metadataAuthId, userType: {%s}, user: {%s}, metadataId: {%s}", userType.String(), user, metadataId)
	return true, nil
}

func RemoveUserMetadataAuthIdByMetadataId(db KeyValueStore, userType apicommonpb.UserType, user, metadataId string) error {
	key := GetUserMetadataAuthByMetadataIdKey(userType, user, metadataId)

	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	log.Debugf("Remove metadataAuthId, userType: {%s}, user: {%s}, metadataId: {%s}", userType.String(), user, metadataId)
	return db.Delete(key)
}

// about metadata history used task.
func StoreMetadataHistoryTaskId(db KeyValueStore, metadataId, taskId string) error {
	// prefix + metadataId + taskId -> index
	item_key := GetMetadataHistoryTaskKey(metadataId, taskId)
	has, err := db.Has(item_key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case nil == err && has:
		return nil   // It have been exists, don't inscrease count
	}

	// When taskId have not by metadata, inscrease metadata used taskId count
	// and put taskId on metadataId mapping.
	//
	// prefix + metadataId -> history task count
	count_key := GetMetadataHistoryTaskCountKey(metadataId)
	count_val, err := db.Get(count_key)

	var count uint32

	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err):
		// do nothing
	case nil == err && len(count_val) != 0:
		count = bytesutil.BytesToUint32(count_val)
	}
	count ++

	count_val = bytesutil.Uint32ToBytes(count)

	// First: put taskId on metadata mapping.
	if err := db.Put(item_key, count_val); nil != err {
		return err
	}
	log.Debugf("InscreaseMetadataHistoryTaskCount, metadataId: {%s}, taskId: {%s}, count: {%d}", metadataId, taskId, count)
	// Second: inscease taskId count by metadata.
	return db.Put(count_key, count_val)
}

func HasMetadataHistoryTaskId (db DatabaseReader,metadataId, taskId string) (bool, error) {

	item_key := GetMetadataHistoryTaskKey(metadataId, taskId)
	has, err := db.Has(item_key)
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err):
		return false, nil
	case nil == err && !has:
		return false, nil
	}
	return true, nil
}

func QueryMetadataHistoryTaskIdCount(db DatabaseReader, metadataId string) (uint32, error) {
	// prefix + metadataId -> history task count
	key := GetMetadataHistoryTaskCountKey(metadataId)
	val, err := db.Get(key)

	var count uint32

	switch {
	case IsNoDBNotFoundErr(err):
		return 0, err
	case IsDBNotFoundErr(err):
		return 0, nil
	case nil == err && len(val) != 0:
		count = bytesutil.BytesToUint32(val)
	}
	return count, nil
}

func QueryMetadataHistoryTaskIds (db KeyValueStore, metadataId string) ([]string, error) {
	// prefix + metadataId + taskId -> index
	prefixAndMetadataId := GetMetadataHistoryTaskKeyPrefixByMetadataId(metadataId)
	it := db.NewIteratorWithPrefixAndStart(prefixAndMetadataId, nil)
	defer it.Release()

	arr := make([]string, 0)
	tmp := make(map[string]struct{}, 0)
	for it.Next() {
		if len(it.Key()) != 0 && len(it.Value()) != 0 {
			// key len == len(prefix) + len([]byte(metadataId)) + len([]byte(taskId))
			taskId := string(it.Key()[len(prefixAndMetadataId):])
			if _, ok := tmp[taskId]; !ok {
				tmp[taskId] = struct{}{}
				arr = append(arr, taskId)
			}
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}

	return arr, nil
}

func StoreTaskUpResultFile(db DatabaseWriter, turf *types.TaskUpResultFile) error {
	key := GetTaskResultFileMetadataIdKey(turf.GetTaskId())
	val, err := rlp.EncodeToBytes(turf)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func QueryTaskUpResultFile(db DatabaseReader, taskId string) (*types.TaskUpResultFile, error) {
	key := GetTaskResultFileMetadataIdKey(taskId)
	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}
	var taskUpResultFile types.TaskUpResultFile
	if err = rlp.DecodeBytes(vb, &taskUpResultFile); nil != err {
		return nil, err
	}
	return &taskUpResultFile, nil
}

func QueryTaskUpResultFileList(db DatabaseIteratee) ([]*types.TaskUpResultFile, error) {

	it := db.NewIteratorWithPrefixAndStart(GetTaskResultFileMetadataIdKeyPrefix(), nil)
	defer it.Release()

	arr := make([]*types.TaskUpResultFile, 0)
	for it.Next() {
		if value := it.Value(); len(value) != 0 {
			var taskUpResultFile types.TaskUpResultFile
			if err := rlp.DecodeBytes(value, &taskUpResultFile); nil != err {
				log.WithError(err).Errorf("Failed to call QueryAllTaskUpResultFile, decode db val failed")
				continue
			}
			arr = append(arr, &taskUpResultFile)
		}
	}

	if len(arr) == 0 {
		return nil, ErrNotFound
	}

	return arr, nil
}

func RemoveTaskUpResultFile(db KeyValueStore, taskId string) error {
	key := GetTaskResultFileMetadataIdKey(taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func StoreTaskPowerPartyIds(db DatabaseWriter, taskId string, powerPartyIds []string) error {
	key := GetTaskPowerPartyIdsKey(taskId)
	val, err := rlp.EncodeToBytes(powerPartyIds)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func QueryTaskPowerPartyIds(db DatabaseReader, taskId string) ([]string, error) {
	key := GetTaskPowerPartyIdsKey(taskId)

	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}
	var powerPartyIds []string
	if err = rlp.DecodeBytes(vb, &powerPartyIds); nil != err {
		return nil, err
	}
	return powerPartyIds, nil
}

func RemoveTaskPowerPartyIds(db KeyValueStore, taskId string) error {
	key := GetTaskPowerPartyIdsKey(taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func StoreTaskPartnerPartyIds(db DatabaseWriter, taskId string, partyIds []string) error {
	key := GetTaskPartnerPartyIdsKey(taskId)
	val, err := rlp.EncodeToBytes(partyIds)
	if nil != err {
		return err
	}
	return db.Put(key, val)
}

func HasTaskPartnerPartyIds(db DatabaseReader, taskId string) (bool, error) {
	key := GetTaskPartnerPartyIdsKey(taskId)

	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return false, err
	case IsDBNotFoundErr(err), !has:
		return false, nil
	}
	return true, nil
}

func QueryTaskPartnerPartyIds(db DatabaseReader, taskId string) ([]string, error) {
	key := GetTaskPartnerPartyIdsKey(taskId)

	vb, err := db.Get(key)
	if nil != err {
		return nil, err
	}
	var partyIdArr []string
	if err = rlp.DecodeBytes(vb, &partyIdArr); nil != err {
		return nil, err
	}
	return partyIdArr, nil
}

func RemoveTaskPartnerPartyId(db KeyValueStore, taskId, partyId string) error {
	key := GetTaskPartnerPartyIdsKey(taskId)
	vb, err := db.Get(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && len(vb) == 0:
		return nil
	}

	var partyIdArr []string
	if err = rlp.DecodeBytes(vb, &partyIdArr); nil != err {
		return err
	}

	for i, id := range partyIdArr {
		if id == partyId {
			partyIdArr = append(partyIdArr[:i], partyIdArr[i+1:]...)
			break
		}
	}
	if len(partyIdArr) == 0 {
		return db.Delete(key)
	}
	vb, err = rlp.EncodeToBytes(partyIdArr)
	if nil != err {
		return err
	}
	return db.Put(key, vb)
}

func RemoveTaskPartnerPartyIds(db KeyValueStore, taskId string) error {
	key := GetTaskPartnerPartyIdsKey(taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func StoreMessageCache(db KeyValueStore, value interface{}) error {
	var (
		key []byte
		val []byte
		err error
	)
	switch v := value.(type) {
	case *types.PowerMsg:

		key = GetPowerMsgKey(v.GetPowerId())
		val, err = proto.Marshal(&libtypes.PowerMsg{
			PowerId:   v.GetPowerId(),
			JobNodeId: v.GetJobNodeId(),
			CreateAt:  v.GetCreateAt(),
		})
		if nil != err {
			return fmt.Errorf("marshal powerMsg failed, %s", err)
		}

	case *types.MetadataMsg:
		key = GetMetadataMsgKey(v.GetMetadataId())
		val, err = proto.Marshal(&libtypes.MetadataMsg{
			MetadataId:      v.GetMetadataId(),
			MetadataSummary: v.GetMetadataSummary(),
			ColumnMetas:     v.GetColumnMetas(),
			CreateAt:        v.GetCreateAt(),
		})
		if nil != err {
			return fmt.Errorf("marshal metadataMsg failed, %s", err)
		}
	case *types.MetadataAuthorityMsg:
		key = GetMetadataAuthMsgKey(v.GetMetadataAuthId())
		val, err = proto.Marshal(&libtypes.MetadataAuthorityMsg{
			MetadataAuthId: v.GetMetadataAuthId(),
			User:           v.GetUser(),
			UserType:       v.GetUserType(),
			Auth:           v.GetMetadataAuthority(),
			Sign:           v.GetSign(),
			CreateAt:       v.GetCreateAt(),
		})
		if nil != err {
			return fmt.Errorf("marshal metadataAuthorityMsg failed, %s", err)
		}
	case *types.TaskMsg:
		key = GetTaskMsgKey(v.GetTaskId())
		val, err = proto.Marshal(&libtypes.TaskMsg{
			Data:          v.GetTaskData(),
			PowerPartyIds: v.GetPowerPartyIds(),
		})
		if nil != err {
			return fmt.Errorf("marshal taskMsg failed, %s", err)
		}
	}
	return db.Put(key, val)
}

func RemovePowerMsg(db KeyValueStore, powerId string) error {
	key := GetPowerMsgKey(powerId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func RemoveAllPowerMsg(db KeyValueStore) error {
	it := db.NewIteratorWithPrefixAndStart(GetPowerMsgKeyPrefix(), nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			db.Delete(key)
		}
	}
	return nil
}

func RemoveMetadataMsg(db KeyValueStore, metadataId string) error {
	key := GetMetadataMsgKey(metadataId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func RemoveAllMetadataMsg(db KeyValueStore) error {
	it := db.NewIteratorWithPrefixAndStart(GetMetadataMsgKeyPrefix(), nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			db.Delete(key)
		}
	}
	return nil
}

func RemoveMetadataAuthMsg(db KeyValueStore, metadataAuthId string) error {
	key := GetMetadataAuthMsgKey(metadataAuthId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func RemoveAllMetadataAuthMsg(db KeyValueStore) error {
	it := db.NewIteratorWithPrefixAndStart(GetMetadataAuthMsgKeyPrefix(), nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			db.Delete(key)
		}
	}
	return nil
}

func RemoveTaskMsg(db KeyValueStore, taskId string) error {
	key := GetTaskMsgKey(taskId)
	has, err := db.Has(key)
	switch {
	case IsNoDBNotFoundErr(err):
		return err
	case IsDBNotFoundErr(err), nil == err && !has:
		return nil
	}
	return db.Delete(key)
}

func RemoveAllTaskMsg(db KeyValueStore) error {
	it := db.NewIteratorWithPrefixAndStart(GetTaskMsgKeyPrefix(), nil)
	defer it.Release()

	for it.Next() {
		if key := it.Key(); len(key) != 0 {
			db.Delete(key)
		}
	}
	return nil
}

func QueryPowerMsgArr(db KeyValueStore) (types.PowerMsgArr, error) {

	it := db.NewIteratorWithPrefixAndStart(GetPowerMsgKeyPrefix(), nil)
	defer it.Release()

	arr := make(types.PowerMsgArr, 0)

	for it.Next() {
		if val := it.Value(); len(val) != 0 {
			var res libtypes.PowerMsg
			if err := proto.Unmarshal(val, &res); nil != err {
				continue
			}
			arr = append(arr, &types.PowerMsg{
				PowerId:   res.GetPowerId(),
				JobNodeId: res.GetJobNodeId(),
				CreateAt:  res.GetCreateAt(),
			})
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func QueryMetadataMsgArr(db KeyValueStore) (types.MetadataMsgArr, error) {
	it := db.NewIteratorWithPrefixAndStart(GetMetadataMsgKeyPrefix(), nil)
	defer it.Release()

	arr := make(types.MetadataMsgArr, 0)

	for it.Next() {
		if val := it.Value(); len(val) != 0 {
			var res libtypes.MetadataMsg
			if err := proto.Unmarshal(val, &res); nil != err {
				continue
			}
			arr = append(arr, &types.MetadataMsg{
				MetadataId:      res.GetMetadataId(),
				MetadataSummary: res.GetMetadataSummary(),
				ColumnMetas:     res.GetColumnMetas(),
				CreateAt:        res.GetCreateAt(),
			})
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func QueryMetadataAuthorityMsgArr(db KeyValueStore) (types.MetadataAuthorityMsgArr, error) {
	it := db.NewIteratorWithPrefixAndStart(GetMetadataAuthMsgKeyPrefix(), nil)
	defer it.Release()

	arr := make(types.MetadataAuthorityMsgArr, 0)

	for it.Next() {
		if val := it.Value(); len(val) != 0 {
			var res libtypes.MetadataAuthorityMsg
			if err := proto.Unmarshal(val, &res); nil != err {
				continue
			}
			arr = append(arr, &types.MetadataAuthorityMsg{
				MetadataAuthId: res.GetMetadataAuthId(),
				User:           res.GetUser(),
				UserType:       res.GetUserType(),
				Auth:           res.GetAuth(),
				Sign:           res.GetSign(),
				CreateAt:       res.GetCreateAt(),
			})
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}

func QueryTaskMsgArr(db KeyValueStore) (types.TaskMsgArr, error) {
	it := db.NewIteratorWithPrefixAndStart(GetMetadataAuthMsgKeyPrefix(), nil)
	defer it.Release()

	arr := make(types.TaskMsgArr, 0)

	for it.Next() {
		if val := it.Value(); len(val) != 0 {
			var res libtypes.TaskMsg
			if err := proto.Unmarshal(val, &res); nil != err {
				continue
			}
			arr = append(arr, &types.TaskMsg{
				Data:          types.NewTask(res.GetData()),
				PowerPartyIds: res.GetPowerPartyIds(),
			})
		}
	}
	if len(arr) == 0 {
		return nil, ErrNotFound
	}
	return arr, nil
}


