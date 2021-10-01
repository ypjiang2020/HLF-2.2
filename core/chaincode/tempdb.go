package chaincode

import (
	"encoding/json"
	"strings"
	"sync"
	"log"

)

type VersionedValue struct {
	txid string
	val  []byte
}

type TempDB struct {
	mutex    sync.Mutex
	Sessions map[string]*SessionDB
}

type WriteSet struct {
	keys []string
	val  [][]byte
}

func (ws *WriteSet) append(key string, val []byte) {
	ws.keys = append(ws.keys, key)
	ws.val = append(ws.val, val)
}

type SessionDB struct {
	session   string
	db        map[string][]byte
	writeSets map[string]*WriteSet
}

func newSessionDB(session string) *SessionDB {
	return &SessionDB{
		session: session,
		db:      map[string][]byte{},
	}
}

func (sdb *SessionDB) Get(key string) *VersionedValue {
	// TODO: do we need to read from writeSets?
	val := sdb.db[key]
	versionedValue := &VersionedValue{}
	err := json.Unmarshal(val, versionedValue)
	if err != nil {
		log.Fatalln("get from session db,", key, versionedValue.txid, err)
	}
	return versionedValue
}

func (sdb *SessionDB) Put(txid, key string, val []byte) {
	// sdb.db[key] = val
	if _, ok := sdb.writeSets[txid]; !ok {
		sdb.writeSets[txid] = &WriteSet{}
	}
	sdb.writeSets[txid].append(key, val)
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
		sdb.db[writeset.keys[i]] = writeset.val[i]
	}
}

func newTempDB() *TempDB {
	res := &TempDB{
		Sessions: map[string]*SessionDB{},
	}
	return res
}

func (tdb *TempDB) Get(key, session string) *VersionedValue {
	if session == "" {
		return nil
	}
	if sdb, ok := tdb.Sessions[session]; ok {
		return sdb.Get(key)
	} else {
		return nil
	}
}

func (tdb *TempDB) Put(key, txid string, val []byte) {
	session := GetSessionFromTxid(txid)
	if session == "" {
		return
	}
	if sdb, ok := tdb.Sessions[session]; ok {
		sdb.Put(txid, key, val)
	} else {
		db := newSessionDB(session)
		db.Put(txid, key, val)
		tdb.Sessions[session] = db
	}
}

func (tdb *TempDB) Rollback(txid string) {
	session := GetSessionFromTxid(txid)
	if session == "" {
		return
	}
	if sdb, ok := tdb.Sessions[session]; ok {
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
	if sdb, ok := tdb.Sessions[session]; ok {
		sdb.Commit(txid)
	} else {
		log.Fatalln("something is wrong with the Temp DB, please check it")
	}
}

func (tdb *TempDB) Prune(txid string) {
	session := GetSessionFromTxid(txid)
	delete(tdb.Sessions, session)
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
