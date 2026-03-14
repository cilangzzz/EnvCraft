package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"tsc/internal/backend_service/handler/migration"
	"tsc/internal/backend_service/interceptor"

	// 导入生成的 docs 包（确保已经执行过 swag init）
	_ "tsc/internal/backend_service/docs" // 替换为你的项目模块路径

	// 导入策略模块以触发 init() 自动注册
	_ "tsc/pkg/util/migration/core/strategies"
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

	// 迁移 API 路由组
	migrationHandler := migration.NewHandler()
	migrationGroup := r.Group("/api/v1/migration")
	{
		// 执行迁移
		migrationGroup.POST("/execute", migrationHandler.Execute)

		// 预览迁移（模拟执行）
		migrationGroup.POST("/dry-run", migrationHandler.DryRun)

		// 回滚迁移
		migrationGroup.POST("/rollback", migrationHandler.Rollback)

		// 导出配置
		migrationGroup.POST("/export", migrationHandler.Export)

		// 导入配置
		migrationGroup.POST("/import", migrationHandler.Import)

		// 获取任务列表
		migrationGroup.GET("/tasks", migrationHandler.ListTasks)

		// 获取任务详情
		migrationGroup.GET("/tasks/:task_id", migrationHandler.GetTask)

		// 获取可用策略列表
		migrationGroup.GET("/strategies", migrationHandler.ListStrategies)
	}
}
