package security

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/types"
	"github.com/thanishsid/dingilink-server/internal/types/apperror"
	"github.com/thanishsid/dingilink-server/internal/types/ctxt"
)

func init() {
	// Load roles to role map
	for _, r := range Roles {
		roleMap[r.Name] = r
	}

	// Load permssions to permission map
	for _, p := range Permissions {
		permissionMap[p.Name] = p
	}
}

func InitSecurity(ctx context.Context, d db.DBQ) error {
	tx, err := d.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := seedPermissions(ctx, tx); err != nil {
		return err
	}

	if err := seedRoles(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

type UserInfo struct {
	Authenticated bool
	User          ContextUser
	Location      types.Point
}

type ContextUser struct {
	ID      int64
	Roles   []Role
	TokenID uuid.UUID
}

type SecurityMeasure interface {
	IsRole() (Role, bool)
	IsPermission() (Permission, bool)
}

func GetUserInfo(ctx context.Context) (UserInfo, error) {
	ui, ok := ctx.Value(ctxt.USER_INFO_CTX_KEY).(UserInfo)
	if !ok {
		return ui, errors.New("user info not found")
	}

	return ui, nil
}

func Authorize(ctx context.Context, measures ...SecurityMeasure) (UserInfo, error) {
	ui, err := GetUserInfo(ctx)
	if err != nil {
		return ui, err
	}

	if !ui.Authenticated {
		return ui, apperror.ErrForbidden
	}

	for _, m := range measures {
		role, isRole := m.IsRole()
		permission, isPermisison := m.IsPermission()

		switch true {
		case isRole:
			hasRole := slices.ContainsFunc(ui.User.Roles, func(r Role) bool {
				return role.Name == r.Name
			})

			if !hasRole {
				return ui, apperror.ErrForbidden
			}
		case isPermisison:
			hasPermission := slices.ContainsFunc(ui.User.Roles, func(r Role) bool {
				return r.HasPermission(permission)
			})

			if !hasPermission {
				return ui, apperror.ErrForbidden
			}
		}
	}

	return ui, nil
}
