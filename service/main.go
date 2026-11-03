package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"service/internal/handler"
	"service/internal/server"
	"service/internal/service"
)

func main() {
	h := handler.New(service.NewIngest(), service.NewChat())

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.NewRouter(h),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("启动服务失败，详细错误信息：", err)
		}
	}()
	log.Println("服务已启动，监听地址：", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("关闭服务失败，详细错误信息：", err)
	}
	log.Println("服务已退出")
}
