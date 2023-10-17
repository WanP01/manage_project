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

// 递归查找
type ProjectMenuChild struct {
	ProjectMenu
	Children []*ProjectMenuChild
}

func CovertChild(allpms []*ProjectMenu) []*ProjectMenuChild {
	var Pmcs []*ProjectMenuChild
	for _, v := range allpms { //先找出位于最顶层就是pid == 0 的menu 加入 list
		if v.Pid == 0 {
			firstPmc := &ProjectMenuChild{}
			copier.Copy(firstPmc, v)
			Pmcs = append(Pmcs, firstPmc)
		}
	}
	FindChildPMC(allpms, Pmcs)
	return Pmcs
}
func FindChildPMC(allpms []*ProjectMenu, Pmcs []*ProjectMenuChild) {
	for _, pmc := range Pmcs {
		for _, pm := range allpms { // 再遍历一遍allpms 里不是顶层节点的 menu，找到对应顶层节点的子节点
			if pm.Pid == pmc.Id {
				childPmc := &ProjectMenuChild{}
				copier.Copy(childPmc, pm)
				pmc.Children = append(pmc.Children, childPmc)
			}
		}
		FindChildPMC(allpms, pmc.Children)
	}
}
