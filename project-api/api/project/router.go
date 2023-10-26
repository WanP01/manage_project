package project

import (
	"log"
	"project-api/api/grpc"
	"project-api/api/middleware"
	"project-api/router"

	"github.com/gin-gonic/gin"
)

// 路由初始化
func init() {
	log.Printf("init project router")
	pu := &RouterProject{}
	router.Register(pu)
}

type RouterProject struct {
}

func (pu *RouterProject) Route(r *gin.Engine) {
	// 初始化Project的Grpc Client=》 ProjectGrpcClinet
	grpc.InitProjectGrpcClient()
	//注册验证码函数
	h := NewHandlerProject()
	group := r.Group("/project")
	group.Use(middleware.TokenVerify())
	group.POST("/index", h.index)
	group.POST("/project/selfList", h.myProjectList)
	group.POST("/project", h.myProjectList)
	group.POST("/project_template", h.projectTemplate)
	group.POST("/project/save", h.projectSave)
	group.POST("/project/read", h.projectRead)
	group.POST("/project/recycle", h.projectRecycle)
	group.POST("/project/recovery", h.projectRecovery)
	group.POST("/project_collect/collect", h.projectCollect)
	group.POST("/project/edit", h.projectEdit)
	group.POST("/project/getLogBySelfProject", h.getLogBySelfProject)

	t := NewHandleTask()
	group.POST("/task_stages", t.taskStages)
	group.POST("project_member/index", t.memberProjectList)
	group.POST("task_stages/tasks", t.taskList)
	group.POST("task/save", t.taskSave)
	group.POST("task/sort", t.taskSort)
	group.POST("/task/selfList", t.myTaskList)
	group.POST("/task/read", t.taskRead)
	group.POST("/task_member", t.listTaskMember)
	group.POST("/task/taskLog", t.taskLog)
	group.POST("/task/_taskWorkTimeList", t.taskWorkTimeList)
	group.POST("/task/saveTaskWorkTime", t.saveTaskWorkTime)
	group.POST("/file/uploadFiles", t.uploadFiles)
	group.POST("/task/taskSources", t.taskSources)
	group.POST("/task/createComment", t.createComment)

	a := NewHandlerAccount()
	group.POST("/account", a.account)

	d := NewHandlerDepartment()
	group.POST("/department", d.department)
	group.POST("/department/save", d.save)
	group.POST("/department/read", d.read)

	auth := NewAuth()
	group.POST("/auth", auth.authList)
}
