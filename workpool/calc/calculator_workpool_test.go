package calc

import (
	"testing"
	"time"
)

func calcAttrTaskFunc() TaskFunc {
	taskFunc := TaskFunc(func(args ...interface{}) error {
		//calc := NewCalculator()
		//calcAttrs(calc)
		//attrs := calc.Evaluation()
		//fmt.Println(attrs)
		//fmt.Printf("do something %d; \n", args[0].(int))
		return nil
	})
	return taskFunc
}

func TestCalculator_Workpool(t *testing.T) {

	//创建一个协程池
	p := NewPool(3, 10)

	//启动协程池p
	p.Start()

	//var wg sync.WaitGroup
	//wg.Add(10)
	//开一个协程 不断的向 Pool 输送task任务
	go func() {
		for i := 1; i < 10; i++ {
			task := NewTask(calcAttrTaskFunc(), i)
			p.pushTask(task)
			//wg.Done()
		}
	}()
	//wg.Wait()
	time.Sleep(20 * time.Second)
	task := NewTask(calcAttrTaskFunc(), 10)
	p.pushTask(task)
	time.Sleep(1 * time.Second)
	go func() {
		for i := 11; i <= 30; i++ {
			task := NewTask(calcAttrTaskFunc(), i)
			p.pushTask(task)
		}
	}()
	time.Sleep(300 * time.Second)
}

func BenchmarkCalculator_Workpool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calc := NewCalculator()
		calcAttrs(calc)
		_ = calc.Evaluation()
	}
}
