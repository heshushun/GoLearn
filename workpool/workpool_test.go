package workpool

import (
	"fmt"
	"sync"
	"testing"
)

func TestWorkPool_Start(t *testing.T) {
	wg := sync.WaitGroup{}
	wp := NewWorkPool(3, 50).Start()
	lenth := 100
	wg.Add(lenth)
	for i := 0; i < lenth; i++ {
		wp.PushTaskFunc(func() {
			defer wg.Done()
			fmt.Print(i, " ")
		})
	}
	wg.Wait()
}
