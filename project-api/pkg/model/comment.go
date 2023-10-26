package model

// CommentReq 请求struct
type CommentReq struct {
	TaskCode string   `form:"taskCode"`
	Comment  string   `form:"comment"`
	Mentions []string `form:"mentions"`
}
