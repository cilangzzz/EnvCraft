package router

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {

	// 注册中间件
	r.Use(
		gin.Logger(),   // 内置日志中间件
		gin.Recovery(), // 异常恢复
	)

	//// 健康检查
	//r.GET("/health", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"status": "OK",
	//		"config": gin.H{
	//			"ip":    serverConfig.IP,
	//			"port":  serverConfig.Port,
	//			"debug": serverConfig.Debug,
	//		},
	//	})
	//})
	//
	//// 示例API
	//r.GET("/api/v1/hello", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "Hello World",
	//		"secret":  serverConfig.SecKey,
	//	})
	//})
}
