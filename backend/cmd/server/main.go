package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"novelflow/backend/internal"
	"novelflow/backend/internal/servicecontext"
	"novelflow/config"
	"novelflow/database/task"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	svc := servicecontext.NewServiceContext()
	defer svc.Close()

	// 启动恢复：将上次残留的 running 任务标记为 failed
	if count, err := task.MarkRunningTasksAsFailed(context.Background(), svc.MongoDB, "server restarted"); err != nil {
		log.Printf("Warning: failed to recover running tasks: %v", err)
	} else if count > 0 {
		log.Printf("Recovered %d task(s) from previous run (marked as failed)", count)
	}

	// 创建 Gin 路由器
	router := gin.Default()

	// 配置路由
	internal.SetupRoutes(svc, router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	addr := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动 HTTP 服务器（goroutine 非阻塞）
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %v, shutting down...", sig)

	// 通知所有 in-flight 生成任务停止
	svc.Cancel()

	// 创建关闭截止时间
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 停止接收新请求，等待活跃连接完成
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 等待所有 in-flight 生成 goroutine 完成
	done := make(chan struct{})
	go func() {
		svc.WG.Wait()
		close(done)
	}()
	select {
	case <-done:
		log.Println("All in-flight generations completed")
	case <-shutdownCtx.Done():
		log.Println("Shutdown deadline exceeded; some generations may be incomplete")
	}

	log.Println("Server exited")
}
