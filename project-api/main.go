package main

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"net/http"
	_ "project-api/api" //初始化"user/api/user & /project"路径下的router file
	"project-api/api/middleware"
	"project-api/config"
	"project-api/router"
	srv "project-common"
)

// 仅负责api 启动和路由
func main() {
	r := gin.Default()

	r.Use(middleware.RequestLog())

	//静态文件映射
	r.StaticFS("/upload", http.Dir("upload"))
	//路由注册
	router.InitRouter(r)

	// pprof 开启后默认访问路径 /debug/pprof
	pprof.Register(r)

	/*// 测试内存泄露的函数（实际运行不需要）
	r.GET("/mem", func(c *gin.Context) {
		// 业务代码运行
		outCh := make(chan int)
		// 每秒起10个goroutine，goroutine会阻塞，不释放内存
		tick := time.Tick(time.Second / 10)
		i := 0
		for range tick {
			i++
			fmt.Println(i)
			alloc1(outCh) // 不停的有goruntine因为outCh堵塞，无法释放
		}
	})*/

	//启动服务
	srv.Run(r, config.AppConf.Sc.Name, config.AppConf.Sc.Addr, nil)
}

/*// 测试内存泄露的函数（实际运行不需要）
func alloc1(outCh chan<- int) {
	go alloc2(outCh)
}

// 一个内层函数
func alloc2(outCh chan<- int) {
	func() {
		defer fmt.Println("alloc-fm exit")
		// 分配内存，假用一下
		buf := make([]byte, 1024*1024*10)
		_ = len(buf)
		fmt.Println("alloc done")

		outCh <- 0 //阻塞
		//return
	}()
}*/
