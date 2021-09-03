/*
 * @Date    : 2021/9/1 18:02
 * @File    : peers.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : peers 抽象注册节点，借助一致性哈希算法选择远程节点
 *
 */

package cache

import pb "GoLearn/opencache/cache/cachepb"

// PickPeer 用于根据传入的 key 选择相应远程节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 回调请求 用于远程节点上拿到数据
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
