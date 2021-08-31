/*
 * @Date    : 2021/8/31 14:53
 * @File    : opencache.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : opencache 负责与外部交互，控制缓存存储和获取的主流程
 *
 */

package cache

import (
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
// 回调函数(callback)，在缓存不存在时，调用这个函数，得到源数据。
// 函数类型实现某一个接口，称之为接口型函数
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
// 一个 Group 可以认为是一个缓存的命名空间
type Group struct {
	name      string // 唯一的名称
	getter    Getter // 缓存未命中时获取源数据的回调(callback)
	mainCache cache  // 一开始实现的并发缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) // 所有的缓存组
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 缓存中存在 直接返回
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 缓存中不存在 load 从数据源中获取
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用 回调函数 获取数据源
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	// 重新将源数据添加到缓存 mainCache 中
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
