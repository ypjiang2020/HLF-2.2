package scheduler

import (
	"fmt"
	"log"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger/kvledger/txmgmt/txmgr"
	"github.com/Yunpeng-J/fabric-protos-go/peer"
	mapset "github.com/deckarep/golang-set"
	"strconv"
	"time"
)

type RWDSet struct {
	txId string
	seq int
	readSet mapset.Set
	writeSet mapset.Set
	deltaSet mapset.Set
}

type Scheduler struct {
	Sessions map[string][]RWDSet
	Keys mapset.Set
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Sessions: map[string][]RWDSet{},
		Keys: mapset.NewThreadUnsafeSet(),
	}
}

func (scheduler *Scheduler) parseAction(action *peer.ChaincodeAction) (readSet, writeSet, deltaSet []interface{}) {
	txRWDset := &rwsetutil.TxRwdSet{}
	if err := txRWDset.FromProtoBytes(action.Results); err != nil {
		panic("Fail to retrieve rwset from txn payload")
	}
	if len(txRWDset.NsRwdSets) > 1 {
		ns := txRWDset.NsRwdSets[1]
		for _, read := range ns.KvRwdSet.Reads {
			readSet = append(readSet, read.GetKey())
			scheduler.Keys.Add(read.GetKey())
		}
		for _, write := range ns.KvRwdSet.Writes {
			writeSet = append(writeSet, write.GetKey())
			scheduler.Keys.Add(write.GetKey())
		}
		for _, delta := range ns.KvRwdSet.Deltas {
			deltaSet = append(deltaSet, delta.GetKey())
			scheduler.Keys.Add(delta.GetKey())
		}
	}
	return
}

func (scheduler *Scheduler) Schedule(action *peer.ChaincodeAction, txId string) {
	session := txmgr.GetSessionFromTxid(txId)
	// we assume that consensus service
	// will order transactions according to the sequence number.
	seq, err := strconv.Atoi(txmgr.GetSeqFromTxid(txId))
	if err != nil {
		panic(fmt.Sprintf("failed to extract sequence from transaction %s", txId))
	}
	rs, ws, ds := scheduler.parseAction(action)
	// TODO:
	// transactions with smaller sequence number may deliver later or never deliver
	// analyze transactions according to their dependency
	// delta depends on write
	// read depends on read
	scheduler.Sessions[session] = append(scheduler.Sessions[session], RWDSet{
		readSet:  mapset.NewThreadUnsafeSetFromSlice(rs),
		writeSet: mapset.NewThreadUnsafeSetFromSlice(ws),
		deltaSet: mapset.NewThreadUnsafeSetFromSlice(ds),
		txId: txId,
		seq: seq,
	})



}

func (scheduler *Scheduler) ProcessBlk() []string {
	now := time.Now()
	defer func(sec int64) {
		log.Printf("ProcessBlk in %d us\n", sec)
	}(time.Since(now).Microseconds())

	var res []string





	// clear schedule
	scheduler.Sessions = map[string][]RWDSet{}
	scheduler.Keys = mapset.NewThreadUnsafeSet()

	return res
}
