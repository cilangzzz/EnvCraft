package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"tsc/internal/backend_service/interceptor"

	// 导入生成的 docs 包（确保已经执行过 swag init）
	_ "tsc/internal/backend_service/docs" // 替换为你的项目模块路径
)

func RegisterRoutes(r *gin.Engine) {
	// 注册中间件
	r.Use(
		gin.Logger(),   // 内置日志中间件
		gin.Recovery(), // 异常恢复
	)
	r.Use(interceptor.AuthMiddleware())

	// 添加 Swagger 路由
	// 注意：如果不需要认证可以访问，可以放在 AuthMiddleware 之前
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 其他路由...
}
