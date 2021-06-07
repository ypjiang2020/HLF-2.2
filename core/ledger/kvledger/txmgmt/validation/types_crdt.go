// +build crdt

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validation

import (
	"encoding/json"
	"fmt"

	"github.com/Yunpeng-J/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/privacyenabledstate"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
)

// block is used to used to hold the information from its proto format to a structure
// that is more suitable/friendly for validation
type block struct {
	num uint64
	txs []*transaction
}

// transaction is used to hold the information from its proto format to a structure
// that is more suitable/friendly for validation
type transaction struct {
	indexInBlock            int
	id                      string
	rwset                   *rwsetutil.TxRwSet
	validationCode          peer.TxValidationCode
	containsPostOrderWrites bool
}

// publicAndHashUpdates encapsulates public and hash updates. The intended use of this to hold the updates
// that are to be applied to the statedb  as a result of the block commit
type publicAndHashUpdates struct {
	publicUpdates *privacyenabledstate.PubUpdateBatch
	hashUpdates   *privacyenabledstate.HashedUpdateBatch
	cache         map[string]*AccountWrapper
}

// newPubAndHashUpdates constructs an empty PubAndHashUpdates
func newPubAndHashUpdates() *publicAndHashUpdates {
	return &publicAndHashUpdates{
		privacyenabledstate.NewPubUpdateBatch(),
		privacyenabledstate.NewHashedUpdateBatch(),
		make(map[string]*AccountWrapper),
	}
}

// containsPvtWrites returns true if this transaction is not limited to affecting the public data only
func (t *transaction) containsPvtWrites() bool {
	for _, ns := range t.rwset.NsRwSets {
		for _, coll := range ns.CollHashedRwSets {
			if coll.PvtRwSetHash != nil {
				return true
			}
		}
	}
	return false
}

// retrieveHash returns the hash of the private write-set present
// in the public data for a given namespace-collection
func (t *transaction) retrieveHash(ns string, coll string) []byte {
	if t.rwset == nil {
		return nil
	}
	for _, nsData := range t.rwset.NsRwSets {
		if nsData.NameSpace != ns {
			continue
		}

		for _, collData := range nsData.CollHashedRwSets {
			if collData.CollectionName == coll {
				return collData.PvtRwSetHash
			}
		}
	}
	return nil
}

type Account struct {
	Type            string
	CustomId        string
	CustomName      string
	SavingsBalance  int
	CheckingBalance int
}

type AccountWrapper struct {
	account  *Account
	txHeight *version.Height
	ns       string
	metadata []byte
	dirty    bool
}

func (u *publicAndHashUpdates) UpdateCache() {
	for key, val := range u.cache {
		// fmt.Println("ethereum, write ", key, val.txHeight)
		marshaledData, err := json.Marshal(val.account)
		if err != nil {
			fmt.Println("ethereum: marshaled error")
		}
		u.publicUpdates.PutValAndMetadata(val.ns, key, marshaledData, val.metadata, val.txHeight)
	}

}

// applyWriteSet adds (or deletes) the key/values present in the write set to the publicAndHashUpdates
func (u *publicAndHashUpdates) applyWriteSet(
	txRWSet *rwsetutil.TxRwSet,
	txHeight *version.Height,
	db *privacyenabledstate.DB,
	containsPostOrderWrites bool,
) error {
	u.publicUpdates.ContainsPostOrderWrites =
		u.publicUpdates.ContainsPostOrderWrites || containsPostOrderWrites
	txops, err := prepareTxOps(txRWSet, txHeight, u, db)
	logger.Debugf("txops=%#v", txops)
	if err != nil {
		return err
	}
	for compositeKey, keyops := range txops {
		// fmt.Println("ethereum, len(metadata)=", len(keyops.metadata))
		if compositeKey.coll == "" {
			ns, key := compositeKey.ns, compositeKey.key
			if keyops.isDelete() {
				u.publicUpdates.Delete(ns, key, txHeight)
			} else {
				// how to merge updates using CRDT
				account := &Account{}
				err := json.Unmarshal(keyops.value, account)
				if err != nil {
					fmt.Println("ethereum: unmarshal Account error")
				}
				// fmt.Println("ethereum: account", account.Type, account.CustomName, account.CheckingBalance)
				if account.Type == "crdt" {
					if ori, ok := u.cache[key]; ok {
						ori.account.CheckingBalance += account.CheckingBalance
						data, err := json.Marshal(ori.account)
						if err != nil {
							fmt.Println("ethereum: marshal erorr", err)
						}
						u.publicUpdates.PutValAndMetadata(ns, key, data, keyops.metadata, txHeight)
					} else {
						val, err := db.VersionedDB.GetState(ns, key)
						if err != nil {
							fmt.Println("ethereum: versionedDB.GetState", ns, key, err)
						}
						ori = &AccountWrapper{
							account: &Account{},
						}
						err = json.Unmarshal(val.Value, ori.account)
						if err != nil {
							fmt.Println("ethereum: unmarshal ori error", err)
						}
						// fmt.Println("ethereum: from db", ori.account.Type, ori.account.CustomName, ori.account.CheckingBalance)
						ori.account.CheckingBalance += account.CheckingBalance
						ori.txHeight = txHeight
						ori.metadata = keyops.metadata
						ori.ns = ns
						u.cache[key] = ori
						data, err := json.Marshal(ori.account)
						if err != nil {
							fmt.Println("ethereum: marshal erorr", err)
						}
						u.publicUpdates.PutValAndMetadata(ns, key, data, keyops.metadata, txHeight)
					}
				} else if account.Type == "account" {
					wrapper := &AccountWrapper{
						account:  account,
						ns:       ns,
						metadata: keyops.metadata,
						txHeight: txHeight,
					}
					u.cache[key] = wrapper
					u.publicUpdates.PutValAndMetadata(ns, key, keyops.value, keyops.metadata, txHeight)
				} else {
					// other normal transactions
					u.publicUpdates.PutValAndMetadata(ns, key, keyops.value, keyops.metadata, txHeight)
				}
			}
		} else {
			ns, coll, keyHash := compositeKey.ns, compositeKey.coll, []byte(compositeKey.key)
			if keyops.isDelete() {
				u.hashUpdates.Delete(ns, coll, keyHash, txHeight)
			} else {
				u.hashUpdates.PutValHashAndMetadata(ns, coll, keyHash, keyops.value, keyops.metadata, txHeight)
			}
		}
	}
	return nil
}
