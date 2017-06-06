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
var keyID uint32
func DataKey(key string) string {
	buffer := make([]byte, 4)
	_, err := rand.Read(buffer)
	backup := atomic.AddUint32(&keyID, 1)
	// 防止过大
	atomic.CompareAndSwapUint32(&keyID, 0x99999999, 0x00000001)
	if err != nil {
		return fmt.Sprintf("%s%08x", key, backup)
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
	n := dataStore.Get(key)
	if n == nil && val != nil { // 新创建
		dataStore.Create(key).SetValue(val)
		return true // change!
	} else if n == nil && val == nil { // 未变更
		return false
	} else if n != nil && val != nil { // 修改
		return n.SetValue(val)
	} else /*if n!= nil && val == nil */ { // 删除
		dataStore.Remove(key)
		return true
	}
}

func DataDel(key string) bool {
	return dataStore.Remove(key) != nil
}

func DataGet(key string, s io.Writer, y int) {
	n := dataStore.Get(key)
	if n == nil {
		DataWrite(s, key, nil, y)
	}else{
		DataWrite(s, key, n.GetValue(), y)
	}
}

func DataList(key string, s io.Writer, y int, cb func()) {
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
	n := dataStore.Get(key)
	if n != nil {
		n.Walk(func(c *trie.Node) bool {
			return cb(key + "/" + c.Key, c.GetValue())
		})
	}
}

func DataCleanup(s *server.Session) {
	s.WalkElement(func(key string) bool {
		dataStore.Remove(key)
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
