package chaincode

import (
	"strings"
	"sync"

	pb "github.com/Yunpeng-J/fabric-protos-go/peer"
)

type TxPool struct {
	queues map[string][]*pb.ChaincodeMessage
	mutex  sync.Mutex
}

func newTxPool() *TxPool {
	res := &TxPool{}
	return res
}

func (tp *TxPool) Put(msg *pb.ChaincodeMessage) {
	temp := strings.Split(msg.Txid, "_")
	// seq := temp[0]
	session := temp[1]
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	if queue, ok := tp.queues[session]; ok {
		// make sure that client send txn synchronously
		queue = append(queue, msg)
	} else {
		queue := []*pb.ChaincodeMessage {msg}
		tp.queues[session] = queue
	}
}
