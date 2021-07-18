package auth

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/dgrijalva/jwt-go/test"
	"github.com/micro/cli/v2"
	"github.com/micro/micro/v2/plugin"
	"log"
	"net/http"
)

// 认证相关参数

// Claims是一些实体（通常指的用户）的状态和额外的元数据
type Claims struct {
	// 在jwt默认Claims基础上增加用户ID信息
	UserId string `json:"userId"`
	jwt.StandardClaims
}

// 这里是我们自己封装的Plugin工厂方法，可以参考官方插件增加一些options参数便于插件的灵活配置
func NewPlugin() plugin.Plugin {
	var pubKey *rsa.PublicKey
	return plugin.NewPlugin(
		// 插件名
		plugin.WithName("auth"),
		// token解码需要用到公钥，这里顺百年演示了如何配置命令行参数
		plugin.WithFlag(
			&cli.StringFlag{
				Name:  "auth_key",
				Usage: "auth key file",
				Value: "./conf/public.key",
			}),
		// 配置插件初始化操作，cli.Context中包含了项目启动参数
		plugin.WithInit(func(ctx *cli.Context) error {
			pubKeyFile := ctx.String("auth_key")
			pubKey = test.LoadRSAPublicKeyFromDisk(pubKeyFile)
			return nil
		}),
		// 配置处理函数，注意与wrapper不同，他的参数是http包的ResponseWriter和Request
		plugin.WithHandler(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var claims Claims
				token, err := request.ParseFromRequest(
					r,
					request.AuthorizationHeaderExtractor,
					func(*jwt.Token) (interface{}, error) {
						return pubKey, nil
					},
					request.WithClaims(&claims),
				)

				if err != nil {
					log.Print("token invalid: ", err.Error())
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				// token.Valid是否成功，取决于jwt中Claims接口定义的Valid() error方法
				// 本例中我们直接使用了默认Claims实现jwt.StandardClaims提供的方法，实际生产中可以根据需要重写
				if token == nil || !token.Valid {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// todo:虽然是有效的token，但并不意味着此用户有权访问所有接口，演示代码省略鉴权细节

				// 从Claims种解析userID并加入Header
				r.Header.Set("userId", claims.UserId)

				// 通过了上述验证后，必须执行下面这一步，保证其他插件和业务代码的执行
				h.ServeHTTP(w, r)
			})
		}),
	)
}
