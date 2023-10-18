package code_gen

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"text/template"
)

func connectMysql() *gorm.DB {
	//配置MySQL连接参数
	username := "root"    //账号
	password := "root"    //密码
	host := "127.0.0.1"   //数据库地址，可以是Ip或者域名
	port := 3306          //数据库端口
	Dbname := "msproject" //数据库名
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
	fmt.Println(dsn)
	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
	return db
}

type Result struct {
	Field string
	Type  string
}
type StructResult struct {
	StructName string
	Result     []*Result
}
type MessageResult struct {
	MessageName string
	Result      []*Result
}

func GenStruct(table string, structName string) {
	db := connectMysql()
	var results []*Result
	db.Raw(fmt.Sprintf("describe %s", table)).Scan(&results)
	for _, v := range results {
		v.Field = Name(v.Field)
		v.Type = getType(v.Type)
	}
	tmpl, err := template.ParseFiles("./struct.tpl")
	log.Println(err)
	sr := StructResult{StructName: structName, Result: results}
	tmpl.Execute(os.Stdout, sr)
}

// Name 解析字段名 Field 使子串单词 首字母大写 name_user =》NameUser
func Name(name string) string {
	var names = name[:]
	isSkip := false
	var sb strings.Builder
	for index, value := range names {
		if index == 0 { // 字段名Field第一个字母大写
			s := names[:index+1]
			s = strings.ToUpper(s)
			sb.WriteString(s)
			continue
		}
		if isSkip {
			isSkip = false
			continue
		}
		//95 下划线  user_name
		// 遇到下划线，取后面的单词，并使其大写（下划线后的单词需要大写）
		if value == 95 {
			s := names[index+1 : index+2]
			s = strings.ToUpper(s)
			sb.WriteString(s)
			isSkip = true
			continue
		} else {
			s := names[index : index+1]
			sb.WriteString(s)
		}
	}
	return sb.String()
}

// getType MySQL =》 Struct 类型转换
func getType(t string) string {
	if strings.Contains(t, "bigint") {
		return "int64"
	}
	if strings.Contains(t, "varchar") {
		return "string"
	}
	if strings.Contains(t, "text") {
		return "string"
	}
	if strings.Contains(t, "tinyint") {
		return "int"
	}
	if strings.Contains(t, "int") &&
		!strings.Contains(t, "tinyint") &&
		!strings.Contains(t, "bigint") {
		return "int"
	}
	if strings.Contains(t, "double") {
		return "float64"
	}
	return ""
}

// GenProtoMessage grpc proto解析
func GenProtoMessage(table string, messageName string) {
	db := connectMysql()
	var results []*Result
	db.Raw(fmt.Sprintf("describe %s", table)).Scan(&results)
	for _, v := range results {
		v.Field = Name(v.Field)
		v.Type = getMessageType(v.Type)
	}
	var fm template.FuncMap = make(map[string]any)
	fm["Add"] = func(v int, add int) int {
		return v + add
	}
	t := template.New("message.tpl")
	t.Funcs(fm)
	tmpl, err := t.ParseFiles("./message.tpl")
	log.Println(err)
	sr := MessageResult{MessageName: messageName, Result: results}
	err = tmpl.Execute(os.Stdout, sr)
	log.Println(err)
}

// getMessageType Mysql type =》 grpc Type
func getMessageType(t string) string {
	if strings.Contains(t, "bigint") {
		return "int64"
	}
	if strings.Contains(t, "varchar") {
		return "string"
	}
	if strings.Contains(t, "text") {
		return "string"
	}
	if strings.Contains(t, "tinyint") {
		return "int32"
	}
	if strings.Contains(t, "int") &&
		!strings.Contains(t, "tinyint") &&
		!strings.Contains(t, "bigint") {
		return "int32"
	}
	if strings.Contains(t, "double") {
		return "double"
	}
	return ""
}
