package endorser

import (
	"github.com/Yunpeng-J/HLF-2.2/core/common/ccprovider"
	pb "github.com/Yunpeng-J/fabric-protos-go/peer"
	"sync"
)

type TxContext struct {
	txParams      *ccprovider.TransactionParams
	input         *pb.ChaincodeInput
	chaincodeName string
	response      *pb.Response
	event         *pb.ChaincodeEvent
	err           error
	ch            chan struct{}
}

type ContextManager struct {
	mutex    sync.Mutex
	contexts map[int]*TxContext
	nextSeq int
}

func NewContextManager() *ContextManager {
	return &ContextManager{
		mutex:    sync.Mutex{},
		contexts: map[int]*TxContext{},
		nextSeq: 0,
	}
}

func contextId(channelID, txID string) string {
	return channelID + txID
}

func (c *ContextManager) Create(seq int, txp *ccprovider.TransactionParams, in *pb.ChaincodeInput, chaincode string) *TxContext {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := &TxContext{
		txParams:      txp,
		input:         in,
		chaincodeName: chaincode,
		response:      nil,
		event:         nil,
		err:           nil,
		ch:            make(chan struct{}),
	}
	c.contexts[seq] = ctx
	return ctx
}

func (c *ContextManager) Get() *TxContext {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if ctx, ok := c.contexts[c.nextSeq]; ok {
		c.nextSeq += 1
		return ctx
	}
	return nil
}

func (c *ContextManager) Delete(seq int) {
	c.mutex.Lock()
	delete(c.contexts, seq)
	c.mutex.Unlock()
}
func (c *ContextManager) Close() {
	// TODO
}
