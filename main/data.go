package main

import (
	"github.com/terrywh/keytracker/trie"
	"github.com/terrywh/keytracker/server"
	"sync/atomic"
	"sync"
	"fmt"
	"io"
	"path"
	"crypto/rand"
)

var dataStore trie.Trie
var dataStoreL *sync.RWMutex

func init() {
	dataStore = trie.NewTrie()
	dataStoreL = &sync.RWMutex{}
}
var keyID int64
func DataKey(key string) string {
	buffer := make([]byte, 5)
	_, err := rand.Read(buffer)
	backup := atomic.AddInt64(&keyID, 1)
	if err != nil {
		return fmt.Sprintf("%s%010x", key, backup)
	}else{
		return fmt.Sprintf("%s%02x", key, buffer)
	}
}
func DataKeyFlat(k string) string {
	k = path.Clean(k)
	if k == "/" {
		return ""
	} else {
		return k
	}
}

func DataSet(key string, val interface{}) bool {
	dataStoreL.Lock()
	defer dataStoreL.Unlock()
	n := dataStore.Get(key)
	if n == nil && val != nil {
		dataStore.Create(key).SetValue(val)
		return true // change!
	} else if n == nil && val == nil {
		return false
	} else if n != nil && val != nil {
		return n.SetValue(val)
	} else /*if n!= nil && val == nil */ {
		dataStore.Remove(key)
		return true
	}
}

func DataGet(key string, s io.Writer) {
	dataStoreL.RLock()
	defer dataStoreL.RUnlock()
	n := dataStore.Get(key)
	if n == nil {
		DataWrite(s, key, nil, 0)
	}else{
		DataWrite(s, key, n.GetValue(), 0)
	}
}

func DataList(key string, s io.Writer, y int, cb func()) {
	dataStoreL.RLock()
	defer dataStoreL.RUnlock()
	n := dataStore.Get(key)
	if n != nil {
		n.Walk(func(c *trie.Node) bool {
			DataWrite(s, key + "/" + c.Key, c.GetValue(), y)
			return true
		})
	}
	if cb != nil {
		cb()
	}
}

func DataWalk(key string, cb func (key string, val interface{}) bool) {
	dataStoreL.RLock()
	defer dataStoreL.RUnlock()
	n := dataStore.Get(key)
	if n != nil {
		n.Walk(func(c *trie.Node) bool {
			return cb(key + "/" + c.Key, c.GetValue())
		})
	}
}

func DataCleanup(s *server.Session) {
	dataStoreL.Lock()
	defer dataStoreL.Unlock()

	s.WalkTag(func(tag interface{}) bool {
		_tag := tag.(Tag)
		if !_tag.IsWatcher {
			dataStore.Remove(_tag.Key)
		}
		return true
	})
}

func DataWrite(s io.Writer, key string, val interface{}, y int) {
	if val == nil {
		fmt.Fprintf(s, "{\"k\":\"%s\",\"v\":null,\"y\":%d}\n", key, y)
		return
	}
	switch val.(type) {
	case float64:
		fmt.Fprintf(s, "{\"k\":\"%s\",\"v\":%v,\"y\":%d}\n", key, val, y)
	case bool:
		fmt.Fprintf(s, "{\"k\":\"%s\",\"v\":%t,\"y\":%d}\n", key, val, y)
	default:
		fmt.Fprintf(s, "{\"k\":\"%s\",\"v\":\"%v\",\"y\":%d}\n", key, val, y)
	}
}
