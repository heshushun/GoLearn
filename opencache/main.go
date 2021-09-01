package main

/*
$ curl "http://localhost:9999/api?key=Tom"
630

$ curl "http://localhost:9999/api?key=kkk"
kkk not exist
*/

import (
	"GoLearn/opencache/cache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 创建一个cacheGroup
func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[DB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动 缓存 服务器
func startCacheServer(addr string, addrs []string, group *cache.Group) {
	// 创建 HTTPPool
	peers := cache.NewHTTPPool(addr)
	// 添加远程节点信息
	peers.SetPeer(addrs...)
	// 注册到group中
	group.RegisterPeers(peers)

	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 启动 API 服务器	(对外交互)
func startAPIServer(apiAddr string, gee *cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", true, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}
