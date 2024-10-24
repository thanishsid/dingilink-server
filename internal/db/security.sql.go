// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: security.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const DeletePermission = `-- name: DeletePermission :exec
DELETE FROM permissions WHERE name = $1
`

func (q *Queries) DeletePermission(ctx context.Context, name string) error {
	_, err := q.db.Exec(ctx, DeletePermission, name)
	return err
}

const DeleteRefreshToken = `-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE id = $1
`

func (q *Queries) DeleteRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	_, err := q.db.Exec(ctx, DeleteRefreshToken, tokenID)
	return err
}

const DeleteRefreshTokensByUserID = `-- name: DeleteRefreshTokensByUserID :exec
DELETE FROM refresh_tokens WHERE user_id = $1
`

func (q *Queries) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, DeleteRefreshTokensByUserID, userID)
	return err
}

const DeleteRole = `-- name: DeleteRole :exec
DELETE FROM roles WHERE name = $1
`

func (q *Queries) DeleteRole(ctx context.Context, name string) error {
	_, err := q.db.Exec(ctx, DeleteRole, name)
	return err
}

const DeleteRolePermissions = `-- name: DeleteRolePermissions :exec
DELETE FROM role_permissions WHERE role_id = (SELECT r.id FROM roles r WHERE r.name = $1)
`

func (q *Queries) DeleteRolePermissions(ctx context.Context, roleName string) error {
	_, err := q.db.Exec(ctx, DeleteRolePermissions, roleName)
	return err
}

const GetPermissions = `-- name: GetPermissions :many
SELECT id, name, description, sort_index FROM permissions
`

func (q *Queries) GetPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := q.db.Query(ctx, GetPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Permission
	for rows.Next() {
		var i Permission
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.SortIndex,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetRefreshToken = `-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, issued_at FROM refresh_tokens WHERE id = $1
`

func (q *Queries) GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, GetRefreshToken, tokenID)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.ExpiresAt,
		&i.IssuedAt,
	)
	return i, err
}

const GetRoles = `-- name: GetRoles :many
SELECT id, name, description, sort_index FROM roles
`

func (q *Queries) GetRoles(ctx context.Context) ([]Role, error) {
	rows, err := q.db.Query(ctx, GetRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Role
	for rows.Next() {
		var i Role
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.SortIndex,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const InsertRefreshToken = `-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (
    id,
    user_id,
    token,
    expires_at,
    issued_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
`

type InsertRefreshTokenParams struct {
	ID        uuid.UUID
	UserID    int64
	Token     string
	ExpiresAt pgtype.Timestamptz
	IssuedAt  pgtype.Timestamptz
}

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error {
	_, err := q.db.Exec(ctx, InsertRefreshToken,
		arg.ID,
		arg.UserID,
		arg.Token,
		arg.ExpiresAt,
		arg.IssuedAt,
	)
	return err
}

const InsertRolePermission = `-- name: InsertRolePermission :exec
INSERT INTO role_permissions (
    role_id,
    permission_id
) VALUES (
    (SELECT r.id FROM roles r WHERE r.name = $1),
    (SELECT p.id FROM permissions p WHERE p.name = $2)

)
`

type InsertRolePermissionParams struct {
	RoleName       string
	PermissionName string
}

func (q *Queries) InsertRolePermission(ctx context.Context, arg InsertRolePermissionParams) error {
	_, err := q.db.Exec(ctx, InsertRolePermission, arg.RoleName, arg.PermissionName)
	return err
}

const UpsertPermission = `-- name: UpsertPermission :exec
INSERT INTO permissions 
    (
        name, 
        description,
        sort_index
    )
VALUES (
    $1, 
    $2,
    $3
)
ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    sort_index = EXCLUDED.sort_index
`

type UpsertPermissionParams struct {
	Name        string
	Description *string
	SortIndex   int64
}

func (q *Queries) UpsertPermission(ctx context.Context, arg UpsertPermissionParams) error {
	_, err := q.db.Exec(ctx, UpsertPermission, arg.Name, arg.Description, arg.SortIndex)
	return err
}

const UpsertRole = `-- name: UpsertRole :exec
INSERT INTO roles 
    (
        name, 
        description,
        sort_index
    )
VALUES (
    $1, 
    $2,
    $3
)
ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    sort_index = EXCLUDED.sort_index
`

type UpsertRoleParams struct {
	Name        string
	Description *string
	SortIndex   int64
}

func (q *Queries) UpsertRole(ctx context.Context, arg UpsertRoleParams) error {
	_, err := q.db.Exec(ctx, UpsertRole, arg.Name, arg.Description, arg.SortIndex)
	return err
}
