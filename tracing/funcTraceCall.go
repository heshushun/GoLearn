package tracing

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

var num = 3

var funcs map[string]interface{}

func Call(m map[string]interface{}, name string, params ... interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func Begin() {
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("A%d", random)
	_, _ = Call(funcs, fnName)
}

func A1() {
	defer funcTrace()()
	//B1()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("B%d", random)
	_, _ = Call(funcs, fnName)
}

func B1() {
	defer funcTrace()()
	//C1()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("C%d", random)
	_, _ = Call(funcs, fnName)
}

func C1() {
	defer funcTrace()()
	D()
}

func A2() {
	defer funcTrace()()
	//B2()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("B%d", random)
	_, _ = Call(funcs, fnName)
}

func B2() {
	defer funcTrace()()
	//C2()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("C%d", random)
	_, _ = Call(funcs, fnName)
}

func C2() {
	defer funcTrace()()
	D()
}

func A3() {
	defer funcTrace()()
	//B3()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("B%d", random)
	_, _ = Call(funcs, fnName)
}

func B3() {
	defer funcTrace()()
	//C3()
	random := rand.Intn(num) + 1
	fnName := fmt.Sprintf("C%d", random)
	_, _ = Call(funcs, fnName)
}

func C3() {
	defer funcTrace()()
	D()
}

func D() {
	defer funcTrace()()
}

func init()  {
	funcs = map[string]interface{}{
		"A1": A1,
		"B1": B1,
		"C1": C1,
		"A2": A2,
		"B2": B2,
		"C2": C2,
		"A3": A3,
		"B3": B3,
		"C3": C3,
	}
	rand.Seed(time.Now().Unix())
}