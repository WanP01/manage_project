package domain

import (
	"project-grpc/project"
	"project-user/internal/rpc"
)

type ProjectRpcDomain struct {
	lc project.ProjectServiceClient
}

func NewProjectRpcDomain() *ProjectRpcDomain {
	return &ProjectRpcDomain{
		lc: rpc.ProjectGrpcClient,
	}
}
