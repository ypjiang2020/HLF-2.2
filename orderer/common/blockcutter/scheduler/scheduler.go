package scheduler

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/Yunpeng-J/fabric-protos-go/peer"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxUniqueKeys = 65563

type Scheduler struct {
	windowSize int

	sessionTxs       map[string][]*TxNode
	sessionFutureTxs map[string]PriorityQueue
	uniqueKeyMap     map[string]uint64
	uniqueKeyCounter uint64
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		windowSize:       128,
		sessionTxs:       map[string][]*TxNode{},
		sessionFutureTxs: map[string]PriorityQueue{},
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
			readSet[index] |= uint64(1) << (key % 64)
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
			writeSet[index] |= uint64(1) << (key % 64)
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

func (scheduler *Scheduler) Schedule(action *peer.ChaincodeAction, txId string) bool {
	temp := strings.Split(txId, "_+=+_")
	seq := -1
	session := ""
	var err error
	if len(temp) == 3 {
		seq, err = strconv.Atoi(temp[0])
		if err != nil {
			panic(fmt.Sprintf("failed to extract sequence number from transaction %s", txId))
		}
		session = temp[1]
	}
	if sessionQueue, ok := scheduler.sessionTxs[session]; ok {
		if len(sessionQueue) == 0 || sessionQueue[len(sessionQueue)-1].seq+1 == seq {
			sessionQueue = append(sessionQueue, NewTxNode(seq, txId, action))
		} else {
			// future transactions
			if pq, ok := scheduler.sessionFutureTxs[session]; ok {
				// window size
				if pq.Len() == 0 || pq[0].seq+scheduler.windowSize >= seq {
					heap.Push(&pq, NewTxNode(seq, txId, action))
				} else {
					// drop this transaction
					return false
				}
			} else {
				pq := PriorityQueue{NewTxNode(seq, txId, action)}
				scheduler.sessionFutureTxs[session] = pq
			}
		}
	} else {
		scheduler.sessionTxs[session] = append(scheduler.sessionTxs[session], NewTxNode(seq, txId, action))
	}
	return true
}

func (scheduler *Scheduler) ProcessBlk() (result []string, invalidTxns []string) {
	now := time.Now()
	defer func(sec int64) {
		log.Printf("ProcessBlk in %d us\n", sec)
		// clear scheduler

	}(time.Since(now).Microseconds())

	var sessionNames []string
	for k, _ := range scheduler.sessionTxs {
		sessionNames = append(sessionNames, k)
	}
	// sort sessionNames to guarantee determinism
	sort.Strings(sessionNames)

	// build node (one node may contain multiple transactions)
	var allNodes []*Node
	numOfNodes := int32(0)
	for _, sessionName := range sessionNames {
		// for each session
		var txs []*TxNode
		var nodes []*Node
		for _, v := range scheduler.sessionTxs[sessionName] {
			txs = append(txs, v)
		}
		if pq, ok := scheduler.sessionFutureTxs[sessionName]; ok {
			for pq.Len() > 0 && pq[0].seq == txs[len(txs)-1].seq+1 {
				txs = append(txs, heap.Pop(&pq).(*TxNode))
			}
		}
		// reverse
		// for i, j := 0, len(txs)-1; i < j; i, j = i+1, j-1 {
		//	txs[i], txs[j] = txs[j], txs[i]
		// }
		for i := 0; i < len(txs); i += 1 {
			//for i := len(txs) - 1; i >= 0; i -= 1 {
			tx := txs[i]
			rs, ws, ds, dep, ok := scheduler.parseAction(tx.action)
			if ok == false {
				continue
			}
			var found []int
			for j := 0; j < len(nodes); j++ {
				if nodes[j] == nil {
					// be nil because of the following merge
					continue
				}
				cur := nodes[j]
				// TODO: optimize
				for _, dp := range dep {
					for _, item := range cur.txids {
						if dp == item {
							found = append(found, j)
						}
					}
				}
			}
			if len(found) == 0 {
				// new node
				node := &Node{
					index:    -1,
					txids:    []string{tx.txid},
					readSet:  rs,
					writeSet: ws,
					deltaSet: ds,
				}
				nodes = append(nodes, node)
			} else {
				// merge nodes
				var node Node
				for _, idx := range found {
					cur := nodes[idx]
					// merge
					node.txids = append(node.txids, cur.txids...)
					for k := uint32(0); k < (maxUniqueKeys / 64); k++ {
						node.readSet[k] |= cur.readSet[k]
						node.writeSet[k] |= cur.writeSet[k]
						node.deltaSet[k] |= cur.deltaSet[k]
					}
					nodes[idx] = nil
				}
				nodes = append(nodes, &node)
			}
		}
		for _, node := range nodes {
			if node != nil {
				node.index = numOfNodes
				allNodes = append(allNodes, node)
				numOfNodes += 1
			}
		}
	}

	// build dependency graph
	graph := make([][]int32, numOfNodes)
	invgraph := make([][]int32, numOfNodes)
	for i := int32(0); i < numOfNodes; i++ {
		graph[i] = make([]int32, numOfNodes)
		invgraph[i] = make([]int32, numOfNodes)
	}
	for i := int32(0); i < numOfNodes; i++ {
		for j := int32(0); j < numOfNodes; j++ {
			if i == j {
				continue
			}
			for k := uint32(0); k < (maxUniqueKeys / 64); k++ {
				if allNodes[i].writeSet[k]&allNodes[j].readSet[k] != 0 {
					graph[i] = append(graph[i], j)
					invgraph[j] = append(graph[j], i)
					break
				}
				if allNodes[i].deltaSet[k]&allNodes[j].readSet[k] != 0 {
					graph[i] = append(graph[i], j)
					invgraph[j] = append(graph[j], i)
					break
				}
			}
		}
	}

	// schedule
	schedule, invSet := scheduler.getSchedule(&graph, &invgraph)
	nodeLen := len(schedule)

	for i := 0; i < nodeLen; i += 1 {
		for _, txid := range allNodes[schedule[nodeLen - 1 - i]].txids {
			result = append(result, txid)
		}
	}
	for i := int32(0); i < numOfNodes; i += 1 {
		if invSet[i] == true {
			for _, txid := range allNodes[i].txids {
				invalidTxns = append(invalidTxns, txid)
			}
		}
	}

	return result, invalidTxns
}

func (scheduler *Scheduler) getSchedule(graph *[][]int32, invgraph *[][]int32) ([]int32, []bool) {
	dagGenerator := NewJohnsonCE(graph)
	invCount, invSet := dagGenerator.Run()
	n := int32(len(*graph))
	visited := make([]bool, n)
	schedule := make([]int32, 0, n-invCount)
	remainintNode := n - invCount
	cur := int32(0)
	for remainintNode != 0 {
		flag := true
		if visited[cur] || invSet[cur] {
			cur = (cur + 1) % n
			continue
		}
		for _, in := range (*invgraph)[cur] {
			if (visited[in] || invSet[in]) == false {
				cur = in
				flag = false
				break
			}
		}
		if flag {
			visited[cur] = true
			remainintNode -= 1
			schedule = append(schedule, cur)
			for _, next := range (*graph)[cur] {
				if (visited[next] || invSet[next]) == false {
					cur = next
					break
				}
			}
		}
	}
	return schedule, invSet
}

func (scheduler *Scheduler) checkDependency(i int32) bool {
	// TODO: check if scheduler.pendingTxns[i]'s dependent transactions exist
	// for _, txid := range scheduler.txReadDep[i] {
	// 	if idx, ok := scheduler.txidMap[txid]; ok {
	// 		// its dependent transaction exists in the current block
	// 		if scheduler.invalidTxs[idx] == true {
	// 			return false
	// 		}
	// 	} else {
	// 		temp := strings.Split(txid, "_+=+_")
	// 		seq := -1
	// 		session := ""
	// 		var err error
	// 		if len(temp) == 3 {
	// 			seq, err = strconv.Atoi(temp[0])
	// 			if err != nil {
	// 				panic(fmt.Sprintf("failed to extract sequence from transaction %s", txid))
	// 			}
	// 			session = temp[1]
	// 		}
	// 		for _, wait := range scheduler.sessionWait[session] {
	// 			if wait == int32(seq) {
	// 				return false
	// 			}
	// 		}
	// 	}
	// }
	return true
}
