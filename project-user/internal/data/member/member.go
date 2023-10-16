package member

type Member struct {
	Id              int64
	Account         string
	Password        string
	Name            string
	Mobile          string
	Realname        string
	CreateTime      int64
	Status          int
	LastLoginTime   int64
	Sex             int
	Avatar          string
	Idcard          string
	Province        int
	City            int
	Area            int
	Address         string
	Description     string
	Email           string
	DingtalkOpenid  string
	DingtalkUnionid string
	DingtalkUserid  string
}

func (*Member) TableName() string {
	return "ms_member"
}

/*CREATE TABLE `ms_member`  (
`id` bigint(0) NOT NULL AUTO_INCREMENT COMMENT '系统前台用户表',
`account` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '用户登陆账号',
`password` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '登陆密码',
`name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '用户昵称',
`mobile` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '手机',
`realname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '真实姓名',
`create_time` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '创建时间',
`status` tinyint(1) NULL DEFAULT 0 COMMENT '状态',
`last_login_time` varchar(30) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '上次登录时间',
`sex` tinyint(0) NULL DEFAULT 0 COMMENT '性别',
`avatar` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '头像',
`idcard` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '身份证',
`province` int(0) NULL DEFAULT 0 COMMENT '省',
`city` int(0) NULL DEFAULT 0 COMMENT '市',
`area` int(0) NULL DEFAULT 0 COMMENT '区',
`address` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '所在地址',
`description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '备注',
`email` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '邮箱',
`dingtalk_openid` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '钉钉openid',
`dingtalk_unionid` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '钉钉unionid',
`dingtalk_userid` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '钉钉用户id',
PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '用户表' ROW_FORMAT = COMPACT;*/
