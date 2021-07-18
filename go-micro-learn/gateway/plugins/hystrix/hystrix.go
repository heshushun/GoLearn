package hystrix

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/coreos/pkg/httputil"
	"github.com/micro/micro/v2/plugin"
	"go-micro-learn/common/util/web"
	"log"
	"net/http"
)

func NewPlugin() plugin.Plugin {
	return plugin.NewPlugin(
		plugin.WithName("hystrix"),
		plugin.WithHandler(
			handler,
		),
	)
}
func handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 配置断路器
		name := r.Method + "-" + r.RequestURI
		config := hystrix.CommandConfig{
			Timeout: 300,
		}
		hystrix.ConfigureCommand(name, config)

		// 增强http.ResponseWriter
		// 利用重写的Write()和WriteHeader()保证只写入一次返回值的特性
		newW := &web.ResponseWriterPlus{
			ResponseWriter: w,
			Status:         http.StatusOK,
			Written:        false,
		}

		// 增强*http.Request
		// 为原有的请求上下文增加一个cancel()函数
		ctx, cancel := context.WithCancel(r.Context())
		newR := r.WithContext(ctx)

		if err := hystrix.Do(name,
			func() error {
				defer cancel()
				h.ServeHTTP(newW, newR)
				return nil
			},
			func(err error) error {
				// 熔断后直接执行cancel()结束调用
				// 执行此操作会看到一条报错日志：http: proxy error: context canceled
				// 因为我们事实上就是通过cancel()强行结束调用，因此属于正常情况
				defer cancel()
				return httputil.WriteJSONResponse(newW, http.StatusBadGateway, web.Fail(err.Error()))
			},
		); err != nil {
			log.Println("hystrix breaker err: ", err)
			return
		}
	})
}
