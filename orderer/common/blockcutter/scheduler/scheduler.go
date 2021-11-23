package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/Yunpeng-J/fabric-protos-go/peer"
	"log"
	"strconv"
	"strings"
	"time"
)

const maxUniqueKeys = 65563

type Scheduler struct {
	blockSize uint32

	txidMap    map[string]int32
	sessionWait map[string][]int32  // TODO: prune
	invalidTxs []bool

	txReadSet  [][]uint64
	txReadDep  [][]string // txReadDep[i] = "session_seq_txid"
	txWriteSet [][]uint64
	txDeltaSet [][]uint64

	uniqueKeyCounter uint32
	uniqueKeyMap     map[string]uint32
	pendingTxns      []string
}

func NewScheduler() *Scheduler {
	blocksize := uint32(1500)
	return &Scheduler{
		blockSize:        blocksize,
		invalidTxs:       make([]bool, blocksize),
		txidMap:          make(map[string]int32),
		sessionWait:       make(map[string][]int32),
		txReadSet:        make([][]uint64, blocksize),
		txReadDep:        make([][]string, blocksize),
		txWriteSet:       make([][]uint64, blocksize),
		txDeltaSet:       make([][]uint64, blocksize),
		uniqueKeyCounter: 0,
		uniqueKeyMap:     make(map[string]uint32),
		pendingTxns:      make([]string, 0),
	}
}

func (scheduler *Scheduler) parseAction(action *peer.ChaincodeAction) (readSet, writeSet, deltaSet []uint64, depSet []string, success bool) {
	txRWDset := &rwsetutil.TxRwdSet{}
	if err := txRWDset.FromProtoBytes(action.Results); err != nil {
		panic("Fail to retrieve rwset from txn payload")
	}
	if len(txRWDset.NsRwdSets) > 1 {
		ns := txRWDset.NsRwdSets[1]
		for _, read := range ns.KvRwdSet.Reads {
			readkey := read.GetKey()
			// readver := read.GetVersion()
			valbytes := read.GetValue()
			var vval ledger.VersionedValue
			err := json.Unmarshal(valbytes, &vval)
			if err != nil {
				panic(fmt.Sprintf("unmarshal error when extract version from value in readset: %v", err))
			}
			depSet = append(depSet, vval.Txid)
			key, ok := scheduler.uniqueKeyMap[readkey]
			if ok == false {
				scheduler.uniqueKeyMap[readkey] = scheduler.uniqueKeyCounter
				key = scheduler.uniqueKeyCounter
				scheduler.uniqueKeyCounter += 1
			}
			if key >= maxUniqueKeys {
				// cut the block, and re-process this transaction in the next block
				return nil, nil, nil, nil, false
			}
			index := key / 64
			readSet[index] |= (uint64(1) << (key % 64))
		}
		for _, write := range ns.KvRwdSet.Writes {
			writekey := write.GetKey()
			key, ok := scheduler.uniqueKeyMap[writekey]
			if ok == false {
				scheduler.uniqueKeyMap[writekey] = scheduler.uniqueKeyCounter
				key = scheduler.uniqueKeyCounter
				scheduler.uniqueKeyCounter += 1
			}
			if key >= maxUniqueKeys {
				// cut the block, and re-process this transaction in the next block
				return nil, nil, nil, nil, false
			}
			index := key / 64
			writeSet[index] |= (uint64(1) << (key % 64))
		}
		for _, delta := range ns.KvRwdSet.Deltas {
			deltakey := delta.GetKey()
			key, ok := scheduler.uniqueKeyMap[deltakey]
			if ok == false {
				scheduler.uniqueKeyMap[deltakey] = scheduler.uniqueKeyCounter
				key = scheduler.uniqueKeyCounter
				scheduler.uniqueKeyCounter += 1
			}
			if key >= maxUniqueKeys {
				// cut the block, and re-process this transaction in the next block
				return nil, nil, nil, nil, false
			}
			index := key / 64
			deltaSet[index] |= uint64(1) << (key % 64)
		}
	}
	return readSet, writeSet, deltaSet, depSet, true
}

func (scheduler *Scheduler) Schedule(action *peer.ChaincodeAction, txId string) {

	// session := txmgr.GetSessionFromTxid(txId)
	// we assume that consensus service
	// will order transactions according to the sequence number.
	// seq, err := strconv.Atoi(txmgr.GetSeqFromTxid(txId))
	// if err != nil {
	// panic(fmt.Sprintf("failed to extract sequence from transaction %s", txId))
	// }

	tid := len(scheduler.pendingTxns)
	rs, ws, ds, depSet, ok := scheduler.parseAction(action)
	if ok {
		scheduler.txReadSet[tid] = rs
		scheduler.txReadDep[tid] = depSet
		scheduler.txWriteSet[tid] = ws
		scheduler.txDeltaSet[tid] = ds
		scheduler.txidMap[txId] = int32(tid)
		scheduler.pendingTxns = append(scheduler.pendingTxns, txId)
	}
	// TODO:
	// transactions with smaller sequence number may deliver later or never deliver
	// analyze transactions according to their dependency
	// delta depends on write
	// read depends on read

	// scheduler.Sessions[session] = append(scheduler.Sessions[session], RWDSet{
	// 	readSet:  mapset.NewThreadUnsafeSetFromSlice(rs),
	// 	writeSet: mapset.NewThreadUnsafeSetFromSlice(ws),
	// 	deltaSet: mapset.NewThreadUnsafeSetFromSlice(ds),
	// 	txId: txId,
	// 	seq: seq,
	// })

}

func (scheduler *Scheduler) processInvalidTxns() {
	// TODO: we assume transactions comes in order for now.
}

func (scheduler *Scheduler) ProcessBlk() (result []string, invalidTxns []string) {
	now := time.Now()
	defer func(sec int64) {
		log.Printf("ProcessBlk in %d us\n", sec)
		// clear scheduler
		scheduler.processInvalidTxns()
		scheduler.uniqueKeyCounter = 0
		scheduler.txidMap = make(map[string]int32)
		scheduler.pendingTxns = make([]string, 0)
		scheduler.uniqueKeyMap = make(map[string]uint32)
		scheduler.invalidTxs = make([]bool, scheduler.blockSize)
		scheduler.txReadSet = make([][]uint64, scheduler.blockSize)
		scheduler.txReadDep = make([][]string, scheduler.blockSize)
		scheduler.txWriteSet = make([][]uint64, scheduler.blockSize)
		scheduler.txDeltaSet = make([][]uint64, scheduler.blockSize)

	}(time.Since(now).Microseconds())

	if len(scheduler.pendingTxns) <= 1 {
		return scheduler.pendingTxns, nil
	}
	n := len(scheduler.pendingTxns)
	var seqs []int
	var sessions []string
	for i := int32(0); i < int32(n); i++ {
		temp := strings.Split(scheduler.pendingTxns[i], "_+=+_")
		seq := -1
		session := ""
		var err error
		if len(temp) == 3 {
			seq, err = strconv.Atoi(temp[0])
			if err != nil {
				panic(fmt.Sprintf("failed to extract sequence from transaction %s", scheduler.pendingTxns[i]))
			}
			session = temp[1]
		}
		seqs = append(seqs, seq)
		sessions = append(sessions, session)
	}
	// build dependency graph
	graph := make([][]int32, n)
	for i := int32(0); i < int32(n); i++ {
		graph[i] = make([]int32, n)
	}

	// 1. for transactions in the same session
	// deal with internal read-write dependency
	for i := int32(0); i < int32(n); i++ {
		// cur := scheduler.pendingTxns[i]
		valid := true
		for _, txid := range scheduler.txReadDep[i] {
			if idx, ok := scheduler.txidMap[txid]; ok {
				// its dependent transaction exists in the current block
				if scheduler.invalidTxs[idx] == true {
					valid = false
					break
				}
			} else {
				// session
				temp := strings.Split(txid, "_+=+_")
				seq := -1
				session := ""
				var err error
				if len(temp) == 3 {
					seq, err = strconv.Atoi(temp[0])
					if err != nil {
						panic(fmt.Sprintf("failed to extract sequence from transaction %s", txid))
					}
					session = temp[1]
				}
				for _, wait := range scheduler.sessionWait[session] {
					if wait == int32(seq) {
						valid = false
						break
					}
				}
			}
			if valid == false {
				break
			}
		}
		if valid {
			for _, txid := range scheduler.txReadDep[i] {
				if idx, ok := scheduler.txidMap[txid]; ok {
					graph[i] = append(graph[i], idx)
				}
			}
		} else {
			scheduler.invalidTxs[i] = true
		}
	}

	// 2. for transactions in different sessions
	// deal with write-read dependency and delta-write dependency
	for i := int32(0); i < int32(n); i++ {
		for j := int32(0); j < int32(n); j++ {
			if i == j || scheduler.invalidTxs[i] || scheduler.invalidTxs[j] {
				continue
			}
			if sessions[i] != "" && sessions[j] != "" && sessions[i] != sessions[j] {
				for k := uint32(0); k < (maxUniqueKeys / 64); k++ {
					if (scheduler.txWriteSet[i][k] & scheduler.txReadSet[j][k]) != 0 {
						// write-read dependency
						// first read then write
						graph[i] = append(graph[i], j)
						break
					}
					if (scheduler.txDeltaSet[i][k] & scheduler.txWriteSet[j][k]) != 0 {
						// delta-write conflict
						// first write then delta
						graph[i] = append(graph[i], j)
						break
					}
				}
			}
		}
	}
	// TODO: schedule
	// update result
	for i, invalid := range scheduler.invalidTxs {
		if invalid {
			invalidTxns = append(invalidTxns, scheduler.pendingTxns[i])
		}
	}
	return result, invalidTxns
}
