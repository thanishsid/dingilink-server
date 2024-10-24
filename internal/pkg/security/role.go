package security

import (
	"context"
	"fmt"

	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/db"
)

func GetRoleByName(name string) Role {
	return roleMap[name]
}

var roleMap = map[string]Role{}

type Role struct {
	Name        string
	Description string
	Permissions []Permission

	permissionMap map[string]struct{}
}

func (r Role) IsRole() (Role, bool) {
	return r, true
}

func (r Role) IsPermission() (Permission, bool) {
	return Permission{}, false
}

func (r Role) HasPermission(p Permission) bool {
	_, ok := r.permissionMap[p.Name]
	return ok
}

func newRole(name, description string, permissions ...Permission) Role {

	permMap := make(map[string]struct{})

	for _, p := range permissions {
		permMap[p.Name] = struct{}{}
	}

	return Role{
		Name:          name,
		Description:   description,
		Permissions:   permissions,
		permissionMap: permMap,
	}
}

var (
	User = newRole(
		"user", "Standard user with basic permissions",
		MANAGE_OWN_POSTS,
	)

	Admin = newRole(
		"admin", "Administrator with extended permissions",
		BAN_USERS,
		MANAGE_OTHERS_POSTS,
	)

	SuperAdmin = newRole(
		"super_admin", "Super Administrator with full permissions",
		CREATE_ADMINISTRATORS,
	)
)

var Roles = []Role{
	User,
	Admin,
	SuperAdmin,
}

func seedRoles(ctx context.Context, d db.Querier) error {
	for idx, role := range Roles {
		if err := d.UpsertRole(ctx, db.UpsertRoleParams{
			Name:        role.Name,
			Description: null.StringFrom(role.Description).Ptr(),
			SortIndex:   int64(idx),
		}); err != nil {
			return fmt.Errorf("failed to insert or update role %s: %w", role.Name, err)
		}

		if err := d.DeleteRolePermissions(ctx, role.Name); err != nil {
			return fmt.Errorf("failed to delete %s role permissions: %w", role.Name, err)
		}

		for _, p := range role.Permissions {
			if err := d.InsertRolePermission(ctx, db.InsertRolePermissionParams{
				RoleName:       role.Name,
				PermissionName: p.Name,
			}); err != nil {
				return fmt.Errorf("failed to insert %s role permissions: %s. %w", role.Name, p.Name, err)
			}
		}

		fmt.Printf("Role %s has been seeded/updated.\n", role.Name)
	}

	return deleteRemovedRoles(ctx, d)
}

// Function to delete roles that were removed from the hardcoded list
func deleteRemovedRoles(ctx context.Context, d db.Querier) error {
	existingRoles, err := d.GetRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch roles: %w", err)
	}

	// Convert hardcoded roles to a map for easy lookup
	rolesMap := make(map[string]struct{})
	for _, role := range Roles {
		rolesMap[role.Name] = struct{}{}
	}

	// Remove roles from DB if they're not in the hardcoded list
	for _, role := range existingRoles {
		if _, exists := rolesMap[role.Name]; !exists {
			if err := d.DeleteRole(ctx, role.Name); err != nil {
				return fmt.Errorf("failed to delete role %s: %w", role.Name, err)
			}

			fmt.Printf("Role %s has been deleted.\n", role.Name)
		}
	}

	return nil
}
