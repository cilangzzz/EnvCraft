package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tsc/internal/backend_service/router"
	"tsc/internal/cfg"
)

func main() {
	// 解析命令行参数
	flag.StringVar(&cfg.GlobalServerConfig.IP, "ip", "0.0.0.0", "服务器监听IP")
	flag.StringVar(&cfg.GlobalServerConfig.Port, "port", "8080", "服务器监听端口")
	flag.StringVar(&cfg.GlobalServerConfig.SecKey, "key", "default-secret-key", "安全密钥")
	flag.BoolVar(&cfg.GlobalServerConfig.Debug, "debug", false, "是否开启调试模式")
	flag.Parse()

	// 设置Gin模式
	if cfg.GlobalServerConfig.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化Gin引擎
	r := gin.Default()
	// 注册路由
	router.RegisterRoutes(r)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.GlobalServerConfig.IP, cfg.GlobalServerConfig.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// 打印启动信息
	printStartupInfo()

	// 优雅启停
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	log.Println("服务器已停止")
}

func printStartupInfo() {
	fmt.Println("========================================")
	fmt.Println("Gin 服务器启动配置:")
	fmt.Printf("监听地址: %s:%s\n", cfg.GlobalServerConfig.IP, cfg.GlobalServerConfig.Port)
	fmt.Printf("安全密钥: %s\n", cfg.GlobalServerConfig.SecKey)
	fmt.Printf("调试模式: %v\n", cfg.GlobalServerConfig.Debug)
	fmt.Println("========================================")
}
