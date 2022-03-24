package functrace

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

var m = sync.Map{}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func printTrace(id uint64, name, typ string) {
	v, _ := m.LoadOrStore(id, 0)
	indent := v.(int)
	if typ == "->" {
		m.Store(id, indent+1)
		indent = indent + 1
	} else if typ == "<-" {
		m.Store(id, indent-1)
	}
	indents := ""
	for i := 0; i < indent; i++ {
		indents += "\t"
	}
	fmt.Printf("g[%02d]:%s%s%s\n", id, indents, typ, name)
}

func Trace() func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	id := getGID()
	fn := runtime.FuncForPC(pc)
	name := fn.Name()

	printTrace(id, name, "->")
	return func() {
		printTrace(id, name, "<-")
	}
}
