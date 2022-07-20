package pkg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtPasswd = []byte("~q1w2e3r45678912")

// JWTAuth 中间件，检查token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("authorization")

		result, claims := CheckJWT(token)
		if result["State"] != 200 {
			c.JSON(http.StatusOK, result)
			c.Abort()
			return
		}
		// 继续交由下一个路由处理,并将解析出的信息传递下去
		c.Set("claims", claims)
	}
}

// JWT 签名结构
type JWT struct {
	SigningKey []byte
}

// 一些常量
var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "(GoFuckyourSelf~!)"
)

// 载荷，可以加一些自己需要的信息
type CustomClaims struct {
	UserID int32
	jwt.StandardClaims
}

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取signKey
func GetSignKey() string {
	return SignKey
}

// 这是SignKey
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析Tokne
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

// 更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	var err error
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	var token *jwt.Token
	if token, err = jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	}); err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", TokenInvalid
}

// 生成令牌
func GenerateToken(userID int32) (string, error) {

	j := &JWT{
		[]byte(SignKey),
	}
	claims := CustomClaims{
		userID,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 100),  // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 6000), // 过期时间 一小时
			Issuer:    "Janly",                         //签名的发行者
		},
	}
	var token string
	var err error
	if token, err = j.CreateToken(claims); err != nil {
		return "", err
	}
	data, err := AesCBCEncrypt([]byte(token), jwtPasswd)

	return base64.StdEncoding.EncodeToString(data), err
}

func GetClaims(c *gin.Context) *CustomClaims {
	return c.MustGet("claims").(*CustomClaims)
}

// GetDataByTime 一个需要token认证的测试接口
func GetDataByTime(c *gin.Context) {
	claims := c.MustGet("claims").(*CustomClaims)
	if claims != nil {
		c.JSON(http.StatusOK, gin.H{
			"State":   200,
			"Code":    "200",
			"Message": "token有效",
			"Data":    claims,
		})
	}
}

func CheckJWT(tokenSrc string) (map[string]interface{}, *CustomClaims) {
	if tokenSrc == "" {
		return gin.H{
			"State":   401,
			"Code":    "ERR_TOKEN_AUTHORIZATION",
			"Message": "Token校验错误,参数缺失",
		}, nil
	}
	var tokenBys []byte
	var err error
	if tokenBys, err = base64.StdEncoding.DecodeString(tokenSrc); err != nil {
		return gin.H{
			"State":   401,
			"Code":    "ERR_TOKEN_AUTHORIZATION",
			"Message": "Token错误",
		}, nil
	}
	if tokenBys, err = AesCBCDncrypt(tokenBys, jwtPasswd); err != nil {
		return gin.H{
			"State":   401,
			"Code":    "ERR_TOKEN_AUTHORIZATION",
			"Message": "Token错误",
		}, nil
	}

	j := NewJWT()
	// parseToken 解析token包含的信息
	claims, err := j.ParseToken(string(tokenBys))

	if err != nil {
		if err == TokenExpired {
			return gin.H{
				"State":   401,
				"Code":    "ERR_TOKEN_AUTHORIZATION",
				"Message": "授权已过期",
			}, nil
		}
		return gin.H{
			"State":   401,
			"Code":    "ERR_TOKEN_AUTHORIZATION",
			"Message": err.Error(),
		}, nil
	}

	return gin.H{
		"State":   200,
		"Code":    "SUC_TOKEN_AUTHORIZATION",
		"Message": "校验成功",
	}, claims
}

//RefreshToken 刷新token
func RefreshToken(c *gin.Context) {
	var err = errors.New("authorization 不能为空")
	token := c.Request.Header.Get("authorization")
	fmt.Println(NewJWT().ParseToken(token))
	if token != "" {
		if token, err = NewJWT().RefreshToken(token); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"State":   200,
				"Code":    "SUC_TOKEN_AUTHORIZATION",
				"Message": "刷新Token成功",
				"Data":    gin.H{"Token": token},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"State":   401,
		"Code":    "ERR_TOKEN_AUTHORIZATION",
		"Message": err.Error(),
	})
}
