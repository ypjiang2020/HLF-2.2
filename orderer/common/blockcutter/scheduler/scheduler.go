package scheduler

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Yunpeng-J/HLF-2.2/core/ledger"
	"github.com/Yunpeng-J/HLF-2.2/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/Yunpeng-J/fabric-protos-go/peer"
)

var maxUniqueKeys = 2048 // 65563

type Scheduler struct {
	windowSize int

	sessionTxs       map[string][]*TxNode
	sessionFutureTxs map[string]*PriorityQueue
	uniqueKeyMap     map[string]int
	uniqueKeyCounter int

	Process_blk_latency int
}

func NewScheduler() *Scheduler {
	maxUniqueKeys, _ := strconv.Atoi(os.Getenv("maxUniqueKeys"))
	log.Println("debug v3, maxUniqueKeys", maxUniqueKeys)
	if maxUniqueKeys == 0 {
		log.Println("debug v3")
		maxUniqueKeys = 1024
	}
	return &Scheduler{
		windowSize:       1024,
		sessionTxs:       map[string][]*TxNode{},
		sessionFutureTxs: map[string]*PriorityQueue{},
	}
}

func (scheduler *Scheduler) parseAction(action *peer.ChaincodeAction) ([]uint64, []uint64, []uint64, []string, bool) {
	readSet := make([]uint64, maxUniqueKeys/64)
	writeSet := make([]uint64, maxUniqueKeys/64)
	deltaSet := make([]uint64, maxUniqueKeys/64)
	var depSet []string
	txRWDset := &rwsetutil.TxRwdSet{}
	if err := txRWDset.FromProtoBytes(action.Results); err != nil {
		panic("Fail to retrieve rwset from txn payload")
	}
	// log.Println("debug v1 length rwd sets ", len(txRWDset.NsRwdSets))
	if len(txRWDset.NsRwdSets) > 1 {
		// log.Println("debug v1 namespace of idx=0", txRWDset.NsRwdSets[0].NameSpace)
		ns := txRWDset.NsRwdSets[1]
		// log.Println("debug v1 namespace of idx=1", ns.NameSpace)
		if ns.NameSpace == "lscc" {
			//
			return nil, nil, nil, nil, true
		}
		for _, read := range ns.KvRwdSet.Reads {
			readkey := read.GetKey()
			// readver := read.GetVersion()
			valbytes := read.GetValue()
			var vval ledger.VersionedValue
			err := json.Unmarshal(valbytes, &vval)
			if err != nil {
				// e.g., when read a non-exist account, this will fail.
				// But it's an expected behavior.
				continue
				// panic(fmt.Sprintf("unmarshal error when extract version from value in readset: %v", err))
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
				log.Println("debug v1 drop transaction because read keys are overflow")
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
				// cut the block, and re-process this transaction in the next
				// block
				log.Println("debug v1 drop transaction because write keys are overflow")
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
				log.Println("debug v1 drop transaction because delta keys are overflow")
				return nil, nil, nil, nil, false
			}
			index := key / 64
			deltaSet[index] |= uint64(1) << (key % 64)
		}
	}
	return readSet, writeSet, deltaSet, depSet, true
}

func (scheduler *Scheduler) Schedule(action *peer.ChaincodeAction, txId string) bool {
	// log.Println("debug v1 unique key length", scheduler.uniqueKeyCounter)
	temp := strings.Split(txId, "_+=+_")
	seq := -1
	session := "unknown"
	var err error
	if len(temp) == 3 {
		seq, err = strconv.Atoi(temp[0])
		if err != nil {
			panic(fmt.Sprintf("failed to extract sequence number from transaction %s", txId))
		}
		session = temp[1]
	}
	if session == "unknown" {
		scheduler.sessionTxs[session] = append(scheduler.sessionTxs[session], NewTxNode(seq, txId, action))
		return true
	}
	// future transactions
	if pq, ok := scheduler.sessionFutureTxs[session]; ok {
		// window size
		if pq.Len() == 0 || (*pq)[0].seq+scheduler.windowSize >= seq {
			// log.Println("debug v1 append to future session", session, seq)
			heap.Push(pq, NewTxNode(seq, txId, action))
		} else {
			// drop this transaction
			// log.Println("debug v1 drop transaction, seq is too high", txId)
			return false
		}
	} else {
		// log.Println("debug v1 create heap, insert transaction", txId)
		pq := PriorityQueue{}
		heap.Push(&pq, NewTxNode(seq, txId, action))
		scheduler.sessionFutureTxs[session] = &pq
	}

	return true
}

func (scheduler *Scheduler) ProcessBlk() (result []string, invalidTxns []string) {
	now := time.Now()
	defer func() {
		sec := time.Since(now).Milliseconds()
		scheduler.Process_blk_latency = int(sec)
		log.Printf("debug v3 ProcessBlk in %d ms\n", sec)
		// clear scheduler
		scheduler.sessionTxs = map[string][]*TxNode{}
		scheduler.uniqueKeyCounter = 0
		scheduler.uniqueKeyMap = map[string]int{}

	}()

	sessionNames := []string{}
	for k, _ := range scheduler.sessionTxs {
		sessionNames = append(sessionNames, k)
	}
	for k, _ := range scheduler.sessionFutureTxs {
		sessionNames = append(sessionNames, k)
	}
	// sort sessionNames to guarantee determinism
	sort.Strings(sessionNames)
	// log.Println("debug v1 session names", sessionNames)
	log.Println("debug v3", time.Since(now).Milliseconds())

	__temp := time.Now()

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
			for pq.Len() > 0 && (len(txs) == 0 || (*pq)[0].seq == txs[len(txs)-1].seq+1) {
				// log.Println("debug v1 length of pq", pq.Len(), (*pq)[0].seq)
				txs = append(txs, heap.Pop(pq).(*TxNode))
				// log.Println("debug v1", txs[len(txs)-1].seq)
			}
			if pq.Len() == 0 {
				delete(scheduler.sessionFutureTxs, sessionName)
			}
		}
		// log.Printf("debug v2 number of transactions %d in session %s\n", len(txs), sessionName)
		// reverse
		// for i, j := 0, len(txs)-1; i < j; i, j = i+1, j-1 {
		//	txs[i], txs[j] = txs[j], txs[i]
		// }
		for i := 0; i < len(txs); i += 1 {
			//for i := len(txs) - 1; i >= 0; i -= 1 {
			tx := txs[i]
			rs, ws, ds, dep, ok := scheduler.parseAction(tx.action)
			if ok == false {
				// log.Println("debug v1 parse ation return false for transaction", tx.txid)
				continue
			}
			// log.Printf("debug v2 tx %s dependency %v\n", tx.txid, dep)
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
			mergeNodes := func() {
				// log.Println("debug v2 merge nodes", i, found)
				node := Node{
					index:    -1,
					txids:    []string{tx.txid},
					readSet:  rs,
					writeSet: ws,
					deltaSet: ds,
				}
				for _, idx := range found {
					cur := nodes[idx]
					// merge
					node.txids = append(node.txids, cur.txids...)
					for k := 0; k < (maxUniqueKeys / 64); k++ {
						node.readSet[k] |= cur.readSet[k]
						node.writeSet[k] |= cur.writeSet[k]
						node.deltaSet[k] |= cur.deltaSet[k]
					}
					nodes[idx] = nil
				}
				// log.Println("debug v2 merge txids:", node.txids)
				nodes = append(nodes, &node)
			}

			if len(found) == 0 {
				// // TODO: delta-read conflict
				// for j := 0; j < len(nodes); j++ {
				// 	if nodes[j] == nil {
				// 		// be nil because of the following merge
				// 		continue
				// 	}
				// 	for k := 0; k < (maxUniqueKeys / 64); k++ {
				// 		if ds[k]&nodes[j].readSet[k] != 0 {
				// 			found = append(found, j)
				// 			break
				// 		}
				// 	}
				// }
				// if len(found) != 0 {
				// 	// merge ndoes
				// 	mergeNodes()
				// } else {
				// 	// new node
				// 	// log.Println("debug v2 new node with txid", tx.txid)
				// 	node := &Node{
				// 		index:    -1,
				// 		txids:    []string{tx.txid},
				// 		readSet:  rs,
				// 		writeSet: ws,
				// 		deltaSet: ds,
				// 	}
				// 	nodes = append(nodes, node)
				// }
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
				mergeNodes()
			}
		}
		for _, node := range nodes {
			if node != nil {
				node.index = numOfNodes
				node.weight = len(node.txids)
				allNodes = append(allNodes, node)
				numOfNodes += 1
			}
		}
	}
	log.Println("debug v3 build node in", time.Since(__temp).Milliseconds())
	__temp = time.Now()

	// log.Println("debug v2 number of nodes", numOfNodes)
	// for i := 0; i < int(numOfNodes); i++ {
	// 	ps := fmt.Sprintf("debug v2: node %d contains:", i)
	// 	for j := 0; j < len(allNodes[i].txids); j++ {
	// 		ps += fmt.Sprintf(" %s", allNodes[i].txids[j])
	// 	}
	// 	log.Println(ps)
	// }
	// build dependency graph
	graph := make([][]int32, numOfNodes)
	invgraph := make([][]int32, numOfNodes)
	for i := int32(0); i < numOfNodes; i++ {
		graph[i] = make([]int32, 0, numOfNodes)
		invgraph[i] = make([]int32, 0, numOfNodes)
	}
	for i := int32(0); i < numOfNodes; i++ {
		for j := int32(0); j < numOfNodes; j++ {
			if i == j {
				continue
			}
			for k := 0; k < (maxUniqueKeys / 64); k++ {
				if allNodes[i].writeSet[k]&allNodes[j].readSet[k] != 0 {
					graph[i] = append(graph[i], j)
					invgraph[j] = append(invgraph[j], i)
					break
				}
				if allNodes[i].deltaSet[k]&allNodes[j].readSet[k] != 0 {
					graph[i] = append(graph[i], j)
					invgraph[j] = append(invgraph[j], i)
					break
				}
			}
		}
	}
	log.Println("debug v3 build graph in", time.Since(__temp).Milliseconds())
	__temp = time.Now()
	log.Printf("debug v2 graph\n")
	for i := int32(0); i < numOfNodes; i++ {
		ps := fmt.Sprintf("debug v2 Node weight %d nodeid %d ", len(allNodes[i].txids), i)
		for j := 0; j < len(graph[i]); j++ {
			ps = ps + fmt.Sprintf("%d ", graph[i][j])
		}
		log.Println(ps)
	}

	// schedule
	schedule, invSet := scheduler.getSchedule(&graph, &invgraph, &allNodes)
	nodeLen := len(schedule)
	// log.Println("debug v3 length of schedule", nodeLen)
	// log.Println("debug v3 invSet", invSet)

	for i := 0; i < nodeLen; i += 1 {
		txids := allNodes[schedule[nodeLen-1-i]].txids
		for j, _ := range txids {
			result = append(result, txids[len(txids)-1-j])
		}
	}
	for i := int32(0); i < numOfNodes; i += 1 {
		if invSet[i] == true {
			txids := allNodes[i].txids
			for j, _ := range txids {
				invalidTxns = append(invalidTxns, txids[len(txids)-1-j])
			}
		}
	}
	log.Println("debug v3 schedule in", time.Since(__temp).Milliseconds())

	return result, invalidTxns
}

func (scheduler *Scheduler) getSchedule(graph *[][]int32, invgraph *[][]int32, allNodes *[]*Node) ([]int32, []bool) {
	// dagGenerator := NewJohnsonCE(graph, allNodes)
	dagGenerator := NewFVS(graph, allNodes)
	invCount, invSet := dagGenerator.Run()
	// log.Printf("debug v2 invalid ndoes: %v", invSet)
	invtemp := "debug v2 Node invalid"
	for i, ok := range invSet {
		if ok {
			invtemp += fmt.Sprintf(" %d", i)
		}
	}
	log.Println(invtemp)
	n := int32(len(*graph))
	visited := make([]bool, n)
	schedule := make([]int32, 0, n-invCount)
	remainingNode := n - invCount
	cur := int32(0)
	for remainingNode != 0 {
		// log.Printf("debug v2: remainingNode %d, cur %d", remainingNode, cur)
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
			remainingNode -= 1
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
