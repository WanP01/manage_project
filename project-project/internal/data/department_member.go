package data

type DepartmentMember struct {
	Id               int64
	DepartmentCode   int64
	OrganizationCode int64
	AccountCode      int64
	JoinTime         int64
	IsPrincipal      int
	IsOwner          int
	Authorize        int64
}

func (*DepartmentMember) TableName() string {
	return "ms_department_member"
}
