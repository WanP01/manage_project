package main

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	srv "project-common"
	"project-project/config"
	"project-project/router"
	"project-project/tracing"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 获取与 jaeger对接好的 provider 接口
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp) //设置为全局provider
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	//路由注册
	router.InitRouter(r)

	// User grpc 连接注册 //初始化rpc调用
	router.InitUserGrpc()

	//gRPC注册
	gc := router.RegisterGrpc()

	//etcd 注冊
	router.RegisterEtcd()

	//初始化kafka生产者
	kwClose := config.InitKafkaWriter()

	//初始化Kafka消费者
	ca := config.NewCacheReader()

	// delete cache （删除缓存保持数据一致性）持续监控是否有缓存一致性
	go ca.DeleteCache()

	//用于grpc 优雅退出
	stop := func() {
		gc.Stop()
		kwClose()
		ca.R.Close()
	}
	srv.Run(r, config.AppConf.Sc.Name, config.AppConf.Sc.Addr, stop)
}
