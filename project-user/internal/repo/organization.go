package repo

import (
	"context"
	"project-user/internal/data/organization"
	"project-user/internal/database"
)

type OrganizationRepo interface {
	SaveOrganization(conn database.DbConn, ctx context.Context, org *organization.Organization) error
	FindOrganizationByMemID(ctx context.Context, memId int64) ([]*organization.Organization, error)
}
