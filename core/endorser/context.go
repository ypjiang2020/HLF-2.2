package endorser

import (
	"sync"

	pb "github.com/Yunpeng-J/fabric-protos-go/peer"
)

type TxContext struct {
	proposal *UnpackedProposal
	response *pb.ProposalResponse
	err      error
	session  string
	ch       chan struct{}
}

type ContextManager struct {
	mutex    sync.Mutex
	contexts map[int]*TxContext
	session  string
	nextSeq  int
}

func NewContextManager() *ContextManager {
	return &ContextManager{
		mutex:    sync.Mutex{},
		contexts: map[int]*TxContext{},
		nextSeq:  0,
	}
}

func contextId(channelID, txID string) string {
	return channelID + txID
}

func (c *ContextManager) SetNext(seq int, session string) {
	c.mutex.Lock()
	c.nextSeq = seq
	c.session = session
	c.mutex.Unlock()
}

func (c *ContextManager) Create(seq int, session string, up *UnpackedProposal) *TxContext {
	// log.Printf("debug v10 context create %d", seq)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := &TxContext{
		proposal: up,
		response: nil,
		err:      nil,
		session:  session,
		ch:       make(chan struct{}),
	}
	c.contexts[seq] = ctx
	return ctx
}

func (c *ContextManager) Get() *TxContext {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if ctx, ok := c.contexts[c.nextSeq]; ok && ctx.session == c.session {
		// log.Printf("debug v10 context get %d", c.nextSeq)
		c.nextSeq += 1
		return ctx
	}
	return nil
}

func (c *ContextManager) Delete(seq int) {
	c.mutex.Lock()
	delete(c.contexts, seq)
	// log.Printf("debug v10 context delete %d", seq)
	c.mutex.Unlock()
}
func (c *ContextManager) Close() {
	// TODO
}
