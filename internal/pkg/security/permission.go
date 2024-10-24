package security

import (
	"context"
	"fmt"

	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/db"
)

func GetPermissionByName(name string) Permission {
	return permissionMap[name]
}

var permissionMap = map[string]Permission{}

type Permission struct {
	Name        string
	Description string
}

func (p Permission) IsRole() (Role, bool) {
	return Role{}, false
}

func (p Permission) IsPermission() (Permission, bool) {
	return p, true
}

func newPermission(name, description string) Permission {
	return Permission{
		Name:        name,
		Description: description,
	}
}

var (
	CREATE_ADMINISTRATORS = newPermission("CREATE_ADMINISTRATORS", "create admin users")
	BAN_USERS             = newPermission("BAN_USERS", "Ban regular users")

	MANAGE_OWN_POSTS    = newPermission("MANAGE_OWN_POSTS", "can create, modify and delete their own posts")
	MANAGE_OTHERS_POSTS = newPermission("MANAGE_OTHERS_POSTS", "can modify and delete others posts")
)

var Permissions = []Permission{
	CREATE_ADMINISTRATORS,
	BAN_USERS,

	MANAGE_OWN_POSTS,
	MANAGE_OTHERS_POSTS,
}

func seedPermissions(ctx context.Context, d db.Querier) error {
	for idx, permission := range Permissions {
		if err := d.UpsertPermission(ctx, db.UpsertPermissionParams{
			Name:        permission.Name,
			Description: null.StringFrom(permission.Description).Ptr(),
			SortIndex:   int64(idx),
		}); err != nil {
			return fmt.Errorf("failed to insert or update permission %s: %w", permission.Name, err)
		}

		fmt.Printf("Permission %s has been seeded/updated.\n", permission.Name)
	}

	return deleteRemovedPermissions(ctx, d)
}

// Function to delete permissions that were removed from the hardcoded list
func deleteRemovedPermissions(ctx context.Context, d db.Querier) error {
	existingPermissions, err := d.GetPermissions(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch permissions: %w", err)
	}

	// Convert hardcoded permissions to a map for easy lookup
	permissionsMap := make(map[string]struct{})
	for _, permission := range Permissions {
		permissionsMap[permission.Name] = struct{}{}
	}

	// Remove permissions from DB if they're not in the hardcoded list
	for _, permission := range existingPermissions {
		if _, exists := permissionsMap[permission.Name]; !exists {
			if err := d.DeletePermission(ctx, permission.Name); err != nil {
				return fmt.Errorf("failed to delete permission %s: %w", permission.Name, err)
			}

			fmt.Printf("Permission %s has been deleted.\n", permission.Name)
		}
	}

	return nil
}
