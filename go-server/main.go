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
	"tsc/pkg/backend_service/router"
	"tsc/pkg/cfg"
)

var (
	serverConfig = cfg.ServerConfig{}
)

func main() {
	// 解析命令行参数
	flag.StringVar(&serverConfig.IP, "ip", "0.0.0.0", "服务器监听IP")
	flag.StringVar(&serverConfig.Port, "port", "8080", "服务器监听端口")
	flag.StringVar(&serverConfig.SecKey, "key", "default-secret-key", "安全密钥")
	flag.BoolVar(&serverConfig.Debug, "debug", false, "是否开启调试模式")
	flag.Parse()
	cfg.GlobalServerConfig = &serverConfig

	// 设置Gin模式
	if serverConfig.Debug {
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
		Addr:         fmt.Sprintf("%s:%s", serverConfig.IP, serverConfig.Port),
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
	fmt.Printf("监听地址: %s:%s\n", serverConfig.IP, serverConfig.Port)
	fmt.Printf("安全密钥: %s\n", serverConfig.SecKey)
	fmt.Printf("调试模式: %v\n", serverConfig.Debug)
	fmt.Println("========================================")
}
