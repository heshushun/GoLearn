package tracing

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"time"
)

const JaegerAddr = "127.0.0.1:6831"

func NewJaegerTracer(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler:&config.SamplerConfig{
			Type:     "const",  // 固定采样
			Param:1,  			// 1=全采样、0=不采样
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			LocalAgentHostPort:  JaegerAddr,
		},
		ServiceName: service,
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("Error: connot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

func spanDemo(req string, ctx context.Context) (response string) {
	// 1. 创建span
	span, _ := opentracing.StartSpanFromContext(ctx, "span_demo")
	defer func() {
		// 4. 设置tag
		span.SetTag("request", req)
		span.SetTag("response", response)
		span.Finish()
	}()

	println(req)
	//2. 模拟耗时
	time.Sleep(time.Second)
	//3. 返回
	response = "spanDemoResponse"
	return
}

func spanDemo2(req string, ctx context.Context) (response string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "span_demo2")
	defer func() {
		span.SetTag("request", req)
		span.SetTag("response", response)
		span.Finish()
	}()

	println(req)
	time.Sleep(time.Second)
	response = "spanDemo2Response"
	return
}