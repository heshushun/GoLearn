package tracing

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"testing"
)

func Test_Tracer(t *testing.T) {
	tracer, closer := initJaeger("jaeger-test-demo")
	defer closer.Close()
	// 创建 root span
	span := tracer.StartSpan("span_root")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	r1 := spanDemo("Hello spanDemo", ctx)
	r2 := spanDemo2("Hello spanDemo2", ctx)
	fmt.Printf("Resp demo: %v \n", r1)
	fmt.Printf("Resp demo2: %v\n", r2)
	span.Finish()
}
