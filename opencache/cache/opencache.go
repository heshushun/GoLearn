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
	pb "GoLearn/opencache/cache/cachepb"
	"GoLearn/opencache/cache/singleflight"
	"fmt"
	"log"
	"sync"
)

// 回调函数(callback)，在缓存不存在时，调用这个函数，得到源数据。
type Getter interface {
	Get(key string) ([]byte, error)
}

// 函数类型实现某一个接口，称之为接口型函数
type GetterFunc func(key string) ([]byte, error)

// 实现接口的方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 一个 Group 可以认为是一个缓存的命名空间
type Group struct {
	name      string              // 唯一的名称
	getter    Getter              // 缓存未命中时获取源数据的回调(callback)
	mainCache cache               // 一开始实现的并发缓存
	peers     PeerPicker          // 用来选择远程节点
	loader    *singleflight.Group // 单次请求，防止缓存击穿
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) // 所有的缓存组
)

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
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 将 实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 从缓存获取数据
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

// 加载数据源
func (g *Group) load(key string) (value ByteView, err error) {
	// 防止相同的key时多次调用getFromPeer，导致缓存击穿
	ret, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 选择节点，若非本机节点，则调用 getFromPeer() 从远程获取
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return ret.(ByteView), nil
	}
	return
}

// 从远程节点 获取数据源
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

// 从本机节点 获取数据源
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
