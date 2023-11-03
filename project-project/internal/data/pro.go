package data

import (
	"project-common/encrypts"
	"project-common/tms"
	"project-project/pkg/model"
)

// Project 数据库类型
type Project struct {
	Id                 int64
	Cover              string
	Name               string
	Description        string
	AccessControlType  int
	WhiteList          string
	Sort               int
	Deleted            int
	TemplateCode       int64
	Schedule           float64
	CreateTime         int64
	OrganizationCode   int64
	DeletedTime        string
	Private            int
	Prefix             string
	OpenPrefix         int
	Archive            int
	ArchiveTime        int64
	OpenBeginTime      int
	OpenTaskPrivate    int
	TaskBoardTheme     string
	BeginTime          int64
	EndTime            int64
	AutoUpdateSchedule int
}

func (p *Project) TableName() string {
	return "ms_project"
}

func ToProjectMap(list []*Project) map[int64]*Project {
	m := make(map[int64]*Project, len(list))
	for _, v := range list {
		m[v.Id] = v
	}
	return m
}

// ProjectMember 数据库类型
type ProjectMember struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   int64
}

func (*ProjectMember) TableName() string {
	return "ms_project_member"
}

// ProjectAndMember 响应类型 project && Member 结果汇总
type ProjectAndMember struct {
	Project
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   int64
	OwnerName   string
	Collected   int
}

func (m *ProjectAndMember) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func (m *Project) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func ToMap(orgs []*ProjectAndMember) map[int64]*ProjectAndMember {
	m := make(map[int64]*ProjectAndMember)
	for _, v := range orgs {
		m[v.Id] = v
	}
	return m
}

// ProjectCollection 数据库struct
type ProjectCollection struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	CreateTime  int64
}

func (*ProjectCollection) TableName() string {
	return "ms_project_collection"
}

// ProjectTemplate 数据库struct
type ProjectTemplate struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       int64
	OrganizationCode int64
	Cover            string
	MemberCode       int64
	IsSystem         int
}

func (*ProjectTemplate) TableName() string {
	return "ms_project_template"
}

// ProjectTemplateAll 响应类型
type ProjectTemplateAll struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       string
	OrganizationCode string
	Cover            string
	MemberCode       string
	IsSystem         int
	TaskStages       []*TaskStagesOnlyName
	Code             string
}

// Convert : 将生成的 TaskStagesOnlyName 以及数据库里的 ProjectTemplate 结构转换为 ProjectTemplateAll
func (pt *ProjectTemplate) Convert(taskStages []*TaskStagesOnlyName) *ProjectTemplateAll {
	organizationCode, _ := encrypts.EncryptInt64(pt.OrganizationCode, model.AESKey)
	memberCode, _ := encrypts.EncryptInt64(pt.MemberCode, model.AESKey)
	code, _ := encrypts.EncryptInt64(int64(pt.Id), model.AESKey)
	pta := &ProjectTemplateAll{
		Id:               pt.Id,
		Name:             pt.Name,
		Description:      pt.Description,
		Sort:             pt.Sort,
		CreateTime:       tms.FormatByMill(pt.CreateTime),
		OrganizationCode: organizationCode,
		Cover:            pt.Cover,
		MemberCode:       memberCode,
		IsSystem:         pt.IsSystem,
		TaskStages:       taskStages,
		Code:             code,
	}
	return pta
}

// ToProjectTemplateIds => ProjectTemplate 模板id列表
func ToProjectTemplateIds(pts []ProjectTemplate) []int {
	var ids []int
	for _, v := range pts {
		ids = append(ids, v.Id)
	}
	return ids
}

//CREATE TABLE `ms_project`  (
//`id` bigint(0) UNSIGNED NOT NULL AUTO_INCREMENT,
//`cover` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '封面',
//`name` varchar(90) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '名称',
//`description` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '描述',
//`access_control_type` tinyint(0) NULL DEFAULT 0 COMMENT '访问控制l类型',
//`white_list` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '可以访问项目的权限组（白名单）',
//`order` int(0) UNSIGNED NULL DEFAULT 0 COMMENT '排序',
//`deleted` tinyint(1) NULL DEFAULT 0 COMMENT '删除标记',
//`template_code` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '项目类型',
//`schedule` double(5, 2) NULL DEFAULT 0.00 COMMENT '进度',
//`create_time` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '创建时间',
//`organization_code` bigint(0) NULL DEFAULT NULL COMMENT '组织id',
//`deleted_time` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '删除时间',
//`private` tinyint(1) NULL DEFAULT 1 COMMENT '是否私有',
//`prefix` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '项目前缀',
//`open_prefix` tinyint(1) NULL DEFAULT 0 COMMENT '是否开启项目前缀',
//`archive` tinyint(1) NULL DEFAULT 0 COMMENT '是否归档',
//`archive_time` bigint(0) NULL DEFAULT NULL COMMENT '归档时间',
//`open_begin_time` tinyint(1) NULL DEFAULT 0 COMMENT '是否开启任务开始时间',
//`open_task_private` tinyint(1) NULL DEFAULT 0 COMMENT '是否开启新任务默认开启隐私模式',
//`task_board_theme` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT 'default' COMMENT '看板风格',
//`begin_time` bigint(0) NULL DEFAULT NULL COMMENT '项目开始日期',
//`end_time` bigint(0) NULL DEFAULT NULL COMMENT '项目截止日期',
//`auto_update_schedule` tinyint(1) NULL DEFAULT 0 COMMENT '自动更新项目进度',
//PRIMARY KEY (`id`) USING BTREE,
//INDEX `project`(`order`) USING BTREE
//) ENGINE = InnoDB AUTO_INCREMENT = 13043 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '项目表' ROW_FORMAT = COMPACT;

//CREATE TABLE `ms_project_member`  (
//`id` bigint(0) NOT NULL AUTO_INCREMENT,
//`project_code` bigint(0) NULL DEFAULT NULL COMMENT '项目id',
//`member_code` bigint(0) NULL DEFAULT NULL COMMENT '成员id',
//`join_time` bigint(0) NULL DEFAULT NULL COMMENT '加入时间',
//`is_owner` bigint(0) NULL DEFAULT 0 COMMENT '拥有者',
//`authorize` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '角色',
//PRIMARY KEY (`id`) USING BTREE,
//UNIQUE INDEX `unique`(`project_code`, `member_code`) USING BTREE
//) ENGINE = InnoDB AUTO_INCREMENT = 37 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '项目-成员表' ROW_FORMAT = COMPACT;
