package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/pkg/jwt"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("openid", claims.OpenID)

		c.Next()
	}
}
