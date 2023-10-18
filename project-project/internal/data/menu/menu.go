package menu

import "github.com/jinzhu/copier"

type ProjectMenu struct {
	Id         int64
	Pid        int64
	Title      string
	Icon       string
	Url        string
	FilePath   string
	Params     string
	Node       string
	Sort       int
	Status     int
	CreateBy   int64
	IsInner    int
	Values     string
	ShowSlider int
}

func (*ProjectMenu) TableName() string {
	return "ms_project_menu"
}

// ProjectMenuChild 递归查找，由数据库struct ProjectMenu 生成
type ProjectMenuChild struct {
	ProjectMenu
	StatusText string
	InnerText  string
	FullUrl    string
	Children   []*ProjectMenuChild
}

func CovertChild(allpms []*ProjectMenu) []*ProjectMenuChild {
	var allpmcs []*ProjectMenuChild
	copier.Copy(&allpmcs, allpms)
	for _, v := range allpmcs { // 给前端需要的三个数据赋值
		v.StatusText = getStatus(v.Status)
		v.InnerText = getInnerText(v.IsInner)
		v.FullUrl = getFullUrl(v.Url, v.Params, v.Values)
	}
	var Pmcs []*ProjectMenuChild
	for _, v := range allpmcs { //先找出位于最顶层就是pid == 0 的menu 加入 list
		if v.Pid == 0 {
			firstPmc := &ProjectMenuChild{}
			copier.Copy(firstPmc, v)
			Pmcs = append(Pmcs, firstPmc)
		}
	}
	FindChildPMC(allpmcs, Pmcs)
	return Pmcs
}
func FindChildPMC(allpmcs []*ProjectMenuChild, Pmcs []*ProjectMenuChild) {
	for _, pmc := range Pmcs {
		for _, pm := range allpmcs { // 再遍历一遍allpms 里不是顶层节点的 menu，找到对应顶层节点的子节点
			if pm.Pid == pmc.Id {
				childPmc := &ProjectMenuChild{}
				copier.Copy(childPmc, pm)
				pmc.Children = append(pmc.Children, childPmc)
			}
		}
		FindChildPMC(allpmcs, pmc.Children)
	}
}

func getFullUrl(url string, params string, values string) string {
	if (params != "" && values != "") || values != "" {
		return url + "/" + values
	}
	return url
}

func getInnerText(inner int) string {
	if inner == 0 {
		return "导航"
	}
	if inner == 1 {
		return "内页"
	}
	return ""
}

func getStatus(status int) string {
	if status == 0 {
		return "禁用"
	}
	if status == 1 {
		return "使用中"
	}
	return ""
}
