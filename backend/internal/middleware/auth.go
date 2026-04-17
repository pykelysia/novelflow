package middleware

import (
	"context"
	"strings"

	"novelflow/backend/internal/response"
	"novelflow/backend/pkg/jwt"
	"novelflow/cache"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtUtil *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证令牌
		claims, err := jwtUtil.ValidateAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Unauthorized(c, "token has expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// 验证是否在黑名单中
		if claims.ID != "" {
			inBlacklist, err := cache.IsJWTInBlacklist(context.Background(), claims.ID)
			if err != nil {
				response.InternalServerError(c, "failed to check token blacklist")
				c.Abort()
				return
			}
			if inBlacklist {
				response.Unauthorized(c, "token has been revoked")
				c.Abort()
				return
			}
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
