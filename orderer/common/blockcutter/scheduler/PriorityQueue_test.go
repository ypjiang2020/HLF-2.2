package scheduler_test

import (
	"container/heap"
	"fmt"
	"github.com/Yunpeng-J/HLF-2.2/orderer/common/blockcutter/scheduler"
	"testing"
)

func Test1(t *testing.T) {
	var pq scheduler.PriorityQueue
	heap.Init(&pq)
	heap.Push(&pq, scheduler.NewTxNode(0, nil, nil))
	heap.Push(&pq, scheduler.NewTxNode(2, nil, nil))
	heap.Push(&pq, scheduler.NewTxNode(1, nil, nil))
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*scheduler.TxNode)
		fmt.Println(item)
	}

}
