/*
 * @Date    : 2021/9/2 17:11
 * @File    : singleflight.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : singleflight 防止缓存击穿，用一个group临时数据结构，通关call加锁来防止同一个key极短并发时间内的多次调用。
 *
 */

package singleflight

import "sync"

// call 代表正在进行中，或已经结束的请求
type call struct {
	wg  sync.WaitGroup // 加sync.WaitGroup锁避免重入
	val interface{}    // 返回的值
	err error          // 返回的错误
}

// call 代表正在进行中，或已经结束的请求
type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// 相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 阻塞，如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁，锁加1
	g.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理

	g.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束，锁减1

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
