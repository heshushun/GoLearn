/*
 * @Date    : 2021/8/31 14:53
 * @File    : http.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : http 分布式缓存需要实现节点间通信，提供被其他节点访问的能力(基于http)
 *
 */

package cache

import (
	pb "GoLearn/opencache/cache/cachepb"
	"GoLearn/opencache/cache/consistenthash"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/opencache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex             // guards peers and httpGetters
	peers       *consistenthash.Map    // 类型是一致性哈希算法的 Map，用来根据具体的 key 选择远程节点
	httpGetters map[string]*httpGetter // 映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 处理http协议请求获得数据源
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 编码 HTTP 响应
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// 塞Peer：实例化了一致性哈希算法，并且添加了传入的节点
func (p *HTTPPool) SetPeer(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 拿Peer：根据一致性哈希算法拿到Peer远程节点（HTTP客户端）
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 一致性哈希算法的 Get() 方法，根据具体的 key，选择节点
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		// 根据远程节点获得对应的 HTTP 客户端
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 确保HTTPPool实现了PeerPicker接口 如果没有实现会报错的
var _ PeerPicker = (*HTTPPool)(nil)

// httpGetter 就对应于 HTTP 客户端。
type httpGetter struct {
	baseURL string
}

// httpGetter 回调请求获取数据源
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	// 解码 HTTP 响应
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

// 确保httpGetter实现了PeerGetter接口 如果没有实现会报错的
var _ PeerGetter = (*httpGetter)(nil)
