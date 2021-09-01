/*
 * @Date    : 2021/8/31 14:53
 * @File    : consistenthash.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : consistenthash 一致性哈希，分布式缓存会有多节点。一致性哈希是一种策略，
			用于如果该节点并没有存储缓存值，如何选择合适的节点数据源获取数据。
 *
*/

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希算法
type Hash func(data []byte) uint32

// Map 是一致性哈希算法的主数据结构
type Map struct {
	hash     Hash           // Hash 函数
	replicas int            // 虚拟节点倍数
	hashKeys []int          // 哈希环 [虚拟hash] Sorted
	hashMap  map[int]string // 虚拟节点与真实节点的映射表, 键是虚拟节点的哈希值，值是真实节点的名称。{虚拟hash: key}
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加 真实的key
func (m *Map) Add(keys ...string) {
	// 每一个真实节点 key，对应创建 m.replicas 个虚拟节点
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点：strconv.Itoa(i) + key 的哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.hashKeys = append(m.hashKeys, hash)
			m.hashMap[hash] = key
		}
	}
	// 哈希环需要排序
	sort.Ints(m.hashKeys)
}

// 删除 真实的key
func (m *Map) Remove(key string) {
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.hashKeys, hash)
		m.hashKeys = append(m.hashKeys[:idx], m.hashKeys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}

// 哈希获取 映射到的真实节点
func (m *Map) Get(searchKey string) string {
	if len(m.hashKeys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(searchKey)))
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	idx := sort.Search(len(m.hashKeys), func(i int) bool {
		return m.hashKeys[i] >= (hash % m.hashKeys[len(m.hashKeys)-1])
	})
	// 虚拟节点key
	hashKey := m.hashKeys[idx%len(m.hashKeys)]

	// 真实key m.hashMap[hashKey]
	return m.hashMap[hashKey]
}
