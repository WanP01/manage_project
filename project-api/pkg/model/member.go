package model

type Member struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
}
