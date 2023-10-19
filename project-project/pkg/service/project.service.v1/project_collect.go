package project_service_v1

import (
	"context"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/project"
	"project-project/internal/data/pro"
	"project-project/pkg/model"
	"strconv"
	"time"
)

func (ps *ProjectService) UpdateCollectProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.CollectProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	memberId := msg.MemberId
	collectType := msg.CollectType
	var err error
	if collectType == "collect" {
		pc := &pro.ProjectCollection{
			ProjectCode: projectCode,
			MemberCode:  memberId,
			CreateTime:  time.Now().UnixMilli(),
		}
		err = ps.projectRepo.SaveProjectCollect(ctx, pc)
	}
	if collectType == "cancer" {
		err = ps.projectRepo.DeleteProjectCollect(ctx, msg.MemberId, projectCode)
	}
	if err != nil {
		zap.L().Error("project UpdateCollectProject SaveProjectCollect error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project.CollectProjectResponse{}, nil
}
