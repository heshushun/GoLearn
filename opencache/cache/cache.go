/*
 * @Date    : 2021/8/31 14:53
 * @File    : cache.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : cache 用来做并发控制的，实例化 lru，封装 get 和 add 方法，并添加互斥锁 mu。
 *
 */

package cache

import (
	"GoLearn/opencache/cache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化,一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。
	// 主要用于提高性能，并减少程序内存要求。
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
