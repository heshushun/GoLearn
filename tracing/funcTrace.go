package tracing

import (
	"bytes"
	"runtime"
	"strconv"
)

var routes = map[uint64][]string{}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func funcTrace() func() {
	// skip为0的时候表示当前所在的函数，即栈顶，1是从栈顶往下数第二个，以此类推，
	// line为执行了所在函数内的哪一行，
	// file为函数所在的文件名，
	// pc是所在函数的指针，
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}
	// fmt.Printf("callerInfo: %v  %v  %v \n", pc, file, line)

	fn := runtime.FuncForPC(pc)  // 获取函数名
	name := fn.Name()
	gID := getGID()
	//fmt.Printf("g[%02d] enter: %s\n", gID, name)
	// 记录
	if nodes, ok := routes[gID]; ok {
		nodes = append(nodes, name)
		routes[gID] = nodes
	}else {
		routes[gID] = []string{name}
	}

	outFn := func() {
		//fmt.Printf("g[%02d] exit: %s\n", gID, name)
		// 记录
		if nodes, ok := routes[gID]; ok {
			nodes = append(nodes, name)
			routes[gID] = nodes
		}else {
			routes[gID] = []string{name}
		}
	}
	return outFn
}