package user

import (
	"log"
	LoginServiceV1 "project-user/pkg/service/login.service.v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 全局变量（方便复用）
var UserGrpcClient LoginServiceV1.LoginServiceClient

func InitUserGrpcClient() {
	conn, err := grpc.Dial(":8881", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	UserGrpcClient = LoginServiceV1.NewLoginServiceClient(conn)
}
