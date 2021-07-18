package main

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/test"
	"go-micro-learn/gateway/plugins/auth"
	"log"
	"time"
)

// 加密token的私钥
var priKey *rsa.PrivateKey

// 生成并打印用户ID为123的token
func main() {
	priKey = test.LoadRSAPrivateKeyFromDisk("./gateway/conf/private.key")
	token, err := GenerateToken("123")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("token: ", token)
	}
}

// 根据用户ID产生token
func GenerateToken(userId string) (string, error) {
	// 设置token有效时间
	nowTime := time.Now()
	expireTime := nowTime.Add(3 * time.Hour)

	claims := auth.Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			// 过期时间
			ExpiresAt: expireTime.Unix(),
			// 指定token发行人
			Issuer: "micro-auth",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// 该方法内部生成签名字符串，再用于获取完整、已签名的token
	token, err := tokenClaims.SignedString(priKey)
	return token, err
}
func ParseToken(token string) (*auth.Claims, error) {

	// 用于解析鉴权的声明，方法内部主要是具体的解码和校验的过程，最终返回*Token
	tokenClaims, err := jwt.ParseWithClaims(token, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return priKey.Public(), nil
	})

	if tokenClaims != nil {
		// 从tokenClaims中获取到Claims对象，并使用断言，将该对象转换为我们自己定义的Claims
		// 要传入指针，项目中结构体都是用指针传递，节省空间。
		if claims, ok := tokenClaims.Claims.(*auth.Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err

}
