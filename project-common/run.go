package common

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func Run(r *gin.Engine, srvName string, addr string, stop func()) {
	// r := gin.Default()
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	//保证下面的优雅启停
	go func() {
		log.Printf("web server %s running in %s \n", srvName, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal)
	//SIGINT 用户发送INTR字符(Ctrl+C)触发 kill -2
	//SIGTERM 结束程序(可以被捕获、阻塞或忽略)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting Down project %s web server...\n", srvName)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	// grpc stop （）优退出
	if stop != nil {
		stop()
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("web server %s Shutdown, cause by %v : \n", srvName, err)
	}
	select {
	case <-ctx.Done():
		log.Println("wait timeout...")
	}
	log.Printf("web server %s stop success...", srvName)
}
