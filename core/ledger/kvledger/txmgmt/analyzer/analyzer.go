/*
Optimistic code
 */
package analyzer

import "github.com/Yunpeng-J/fabric-protos-go/ledger/rwset/kvrwset"

type Analyzer struct {

}

func (analyzer *Analyzer) CalculateDeltaFromRWset(readSet []*kvrwset.KVRead, writeSet []*kvrwset.KVWrite) (delta []*kvrwset.KVDelta) {
	return nil
}
