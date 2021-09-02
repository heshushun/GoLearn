/*
 * @Date    : 2021/9/2 20:14
 * @File    : singleflight2.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    :
 *
 */

package singleflight

type result struct {
	val interface{}
	err error
}

type entry struct {
	res   result
	ready chan struct{}
}

type request struct {
	key      string
	fn       func() (interface{}, error)
	response chan result
}

type Group2 struct {
	requests chan request
}

func New() *Group2 {
	g := &Group2{make(chan request)}
	go g.serve()
	return g
}

func (g *Group2) Close() { close(g.requests) }

func (g *Group2) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// Create request for each key
	req := request{key, fn, make(chan result)}
	// Send to g.serve handle
	g.requests <- req
	// Wait for response
	ret := <-req.response
	return ret.val, ret.err
}

func (g *Group2) serve() {
	// Cache the results of each key
	cache := make(map[string]*entry)
	// handle each request
	for r := range g.requests {
		if e, ok := cache[r.key]; !ok {
			e := &entry{
				ready: make(chan struct{}),
			}
			cache[r.key] = e
			go e.call(r)
		} else {
			go e.deliver(r.response)
		}
		//I didn't implement a good way to delete the cache
	}
}

func (e *entry) call(req request) {
	e.res.val, e.res.err = req.fn()
	req.response <- e.res
	close(e.ready)
}

func (e *entry) deliver(resp chan<- result) {
	<-e.ready
	resp <- e.res
}
