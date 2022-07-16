package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 基于 errgroup 实现一个 http server 的启动和关闭 ，
// 以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
func main() {
	g, ctx := errgroup.WithContext(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world!"))
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8099",
	}

	// g1 server 主线程
	g.Go(func() error {
		err := server.ListenAndServe() // 服务启动后会阻塞, 虽然使用的是 go 启动，但是由于 g.WaitGroup 试得其是个阻塞的 协程
		if err != nil {
			log.Println("g1 error,will exit.", err.Error())
		}
		return err
	})

	// g2 停止 server
	g.Go(func() error {
		select {
		case <-ctx.Done():
			log.Println("g2 errgroup exit...")
		}
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		// 这里不是必须的，但是如果使用 _ 的话静态扫描工具会报错，加上也无伤大雅
		defer cancel()

		err := server.Shutdown(timeoutCtx)
		log.Println("shutting down server...")
		log.Printf("g2: %s", err)
		return err
	})

	// g3 接收 linux signal 信号
	g.Go(func() error {
		quit := make(chan os.Signal, 0)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			log.Println("g3, ctx execute cancel...")
			log.Println("g3 error,", ctx.Err().Error())
			return ctx.Err()
		case sig := <-quit:
			return fmt.Errorf("g3 get os signal: %v", sig)
		}
	})

	// g.Wait 等待所有 go执行完毕后执行
	// g.Wait 仅获取最先返回的错误
	fmt.Printf("end, errgroup exiting, %+v\n", g.Wait())
}
