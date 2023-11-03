package rpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/department"
	"project-grpc/menu"
	"project-grpc/project"
	"project-grpc/task"
	"project-user/config"
)

// ProjectGrpcClient 全局变量（方便复用）
var ProjectGrpcClient project.ProjectServiceClient
var TaskGrpcClient task.TaskServiceClient
var AccountGrpcClient account.AccountServiceClient
var DepartmentGrpcClient department.DepartmentServiceClient
var AuthGrpcClient auth.AuthServiceClient
var MenuGrpcClient menu.MenuServiceClient

func InitProjectGrpcClient() {
	// 注册grpc resolver 解析器解析 URL
	etcdRegister := discovery.NewResolver(config.AppConf.Ec.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	// 连接GRPC端口
	conn, err := grpc.Dial("etcd:///project",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	ProjectGrpcClient = project.NewProjectServiceClient(conn)
	TaskGrpcClient = task.NewTaskServiceClient(conn)
	AccountGrpcClient = account.NewAccountServiceClient(conn)
	DepartmentGrpcClient = department.NewDepartmentServiceClient(conn)
	AuthGrpcClient = auth.NewAuthServiceClient(conn)
	MenuGrpcClient = menu.NewMenuServiceClient(conn)
}
