package txmgr

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Yunpeng-J/HLF-2.2/core/ledger"
)

type TempDB struct {
	mutex    sync.Mutex
	Sessions map[string]*SessionDB
}

type WriteSet struct {
	keys []string
	vals [][]byte
}

func (ws *WriteSet) append(key string, val []byte) {
	ws.keys = append(ws.keys, key)
	ws.vals = append(ws.vals, val)
}

type SessionDB struct {
	session   string
	db        map[string][]byte    // key is key
	writeSets map[string]*WriteSet // key is txid
}

func NewTempDB() *TempDB {
	return newTempDB()
}

func newSessionDB(session string) *SessionDB {
	return &SessionDB{
		session:   session,
		db:        map[string][]byte{},
		writeSets: map[string]*WriteSet{},
	}
}

func (sdb *SessionDB) Contain(key string) bool {
	_, ok := sdb.db[key]
	return ok
}

func (sdb *SessionDB) Get(key string) *ledger.VersionedValue {
	// TODO: do we need to read from writeSets?
	val, ok := sdb.db[key]
	if !ok {
		return nil
	}

	versionedValue := &ledger.VersionedValue{}
	err := json.Unmarshal(val, versionedValue)
	if err != nil {
		log.Fatalln("get from session db,", key, versionedValue.Txid, err)
	}
	return versionedValue
}

func (sdb *SessionDB) Put(txid, key string, val []byte) {
	if _, ok := sdb.writeSets[txid]; !ok {
		sdb.writeSets[txid] = &WriteSet{}
	}
	sdb.writeSets[txid].append(key, val)
}

func (sdb *SessionDB) Delete(key string) {
	delete(sdb.db, key)
}

func (sdb *SessionDB) Rollback(txid string) {
	delete(sdb.writeSets, txid)
}

func (sdb *SessionDB) Commit(txid string) {
	if _, ok := sdb.writeSets[txid]; !ok {
		log.Fatalln("write set not found", txid)
	}
	writeset := sdb.writeSets[txid]
	cnt := len(writeset.keys)
	for i := 0; i < cnt; i++ {
		var val = writeset.vals[i]
		// verval := &VersionedValue{
		// 	Txid: txid,
		// 	Val:  writeset.vals[i],
		// }
		// val, err := json.Marshal(verval)
		// if err != nil {
		// 	log.Fatalf("commit to session db, marshal %v", err)

		// }
		sdb.db[writeset.keys[i]] = val
	}
	delete(sdb.writeSets, txid)
}

func newTempDB() *TempDB {
	res := &TempDB{
		Sessions: map[string]*SessionDB{},
	}
	return res
}

func (tdb *TempDB) Get(key, session string) *ledger.VersionedValue {
	if session == "" {
		return nil
	}
	tdb.mutex.Lock()
	sdb, ok := tdb.Sessions[session]
	tdb.mutex.Unlock()
	if ok {
		return sdb.Get(key)
	} else {
		return nil
	}
}

func (tdb *TempDB) Put(key, txid string, val []byte) {
	// val has already been encoded by marshaling (txid, ori_val)
	session := GetSessionFromTxid(txid)
	if session == "" {
		return
	}
	tdb.mutex.Lock()
	sdb, ok := tdb.Sessions[session]
	if ok {
		tdb.mutex.Unlock()
		sdb.Put(txid, key, val)
	} else {
		db := newSessionDB(session)
		tdb.Sessions[session] = db
		tdb.mutex.Unlock()
		db.Put(txid, key, val)
	}
}

func (tdb *TempDB) Rollback(txid string) {
	session := GetSessionFromTxid(txid)
	if session == "" {
		return
	}
	tdb.mutex.Lock()
	sdb, ok := tdb.Sessions[session]
	tdb.mutex.Unlock()
	if ok {
		sdb.Rollback(txid)
	} else {
		log.Fatalln("something is wrong with the Temp DB, please check it")
	}
}
func (tdb *TempDB) Commit(txid string) {
	session := GetSessionFromTxid(txid)
	if session == "" {
		return
	}
	tdb.mutex.Lock()
	sdb, ok := tdb.Sessions[session]
	tdb.mutex.Unlock()
	if ok {
		sdb.Commit(txid)
	} else {
		log.Fatalln("something is wrong with the Temp DB, please check it")
	}
}

// Prune: delete all obsolete keys
func (tdb *TempDB) Prune(ks map[string]string) {
	// TODO: optimization
	tdb.mutex.Lock()
	defer tdb.mutex.Unlock()
	for k, s := range ks {
		if _, ok := tdb.Sessions[s]; !ok {
			// from other sessions
			for _, sdb := range tdb.Sessions {
				sdb.Delete(k)
			}
		}
	}
}

func (tdb *TempDB) String() string {
	var res string
	tdb.mutex.Lock()
	defer tdb.mutex.Unlock()
	for session, sessiondb := range tdb.Sessions {
		res += fmt.Sprintf("session:%s\n", session)
		res += fmt.Sprintf("\tcommitdb:\n")
		for key, val := range sessiondb.db {
			var verval ledger.VersionedValue
			err := json.Unmarshal(val, &verval)
			if err != nil {
				log.Fatalf("stringfy tempdb, unmarshal error: %v", err)

			}
			res += fmt.Sprintf("\t\tkey=%s; txid=%s, val=%s\n", key, verval.Txid, verval.Val)
		}
		res += fmt.Sprintf("\tuncommitdb:\n")
		for txid, rws := range sessiondb.writeSets {
			res += fmt.Sprintf("\t\ttxid=%s:\n", txid)
			for i := 0; i < len(rws.keys); i++ {
				res += fmt.Sprintf("\t\t\tkey=%s; val=%s\n", rws.keys[i], string(rws.vals[i]))
			}
		}
	}
	res += fmt.Sprintf("\n")
	return res
}

// txid format: seqNumber_Session_oriTxid
func GetSessionFromTxid(txid string) string {
	temp := strings.Split(txid, "_+=+_")
	if len(temp) == 1 {
		return ""
	} else if len(temp) == 3 {
		return temp[1]
	}
	return ""
}

func GetSeqFromTxid(txid string) string {
	temp := strings.Split(txid, "_+=+_")
	if len(temp) == 1 {
		return ""
	} else if len(temp) == 3 {
		return temp[0]
	}
	return ""
}
