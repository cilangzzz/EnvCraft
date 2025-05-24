package interceptor

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tsc/internal/cfg"
)

const SECURITY_KEY_HEADER = "SECURITY_KEY_HEADER"
const SECURITY_KEY_QUERY = "security-key"

// AuthMiddleware 验证请求中的密钥
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从配置中获取密钥
		apiKey := cfg.GlobalServerConfig.SecKey
		// 检查请求头中的密钥
		clientKey := c.GetHeader(SECURITY_KEY_HEADER)
		if clientKey == "" {
			// 检查查询参数中的密钥
			clientKey = c.Query(SECURITY_KEY_QUERY)
		}

		if clientKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效的 API 密钥",
			})
			return
		}

		// 密钥验证通过，继续处理请求
		c.Next()

	}
}
