/*
 * @Date    : 2021/8/31 14:53
 * @File    : byteView.go
 * @Version : 1.0.0
 * @Author  : hss
 * @Note    : ByteView 是一个只读数据结构，用来表示缓存值的抽象和封装。
 *
 */

package cache

// 选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
type ByteView struct {
	b []byte // 缓存值, 只读
}

func (v ByteView) Len() int {
	return len(v.b)
}

// 只读，返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
