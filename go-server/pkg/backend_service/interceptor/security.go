package interceptor

import (
	"net/http"
	"tsc/pkg/cfg"
)

// AuthMiddleware 验证请求中的密钥
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 serverConfig 获取密钥
		apiKey := cfg.GlobalServerConfig.SecKey
		if apiKey == "" {
			http.Error(w, "未配置 API 密钥", http.StatusInternalServerError)
			return
		}

		// 检查请求头中的密钥
		clientKey := r.Header.Get("X-API-Key")
		if clientKey == "" {
			// 检查查询参数中的密钥
			clientKey = r.URL.Query().Get("api_key")
		}

		if clientKey != apiKey {
			http.Error(w, "无效的 API 密钥", http.StatusUnauthorized)
			return
		}

		// 密钥验证通过，继续处理请求
		next.ServeHTTP(w, r)
	})
}
