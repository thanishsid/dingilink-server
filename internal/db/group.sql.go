// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: group.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const GetBatchedGroupMembers = `-- name: GetBatchedGroupMembers :many
SELECT 
    gm.id, gm.group_id, gm.user_id, gm.joined_at, gm.is_admin,
    (g.created_by = gm.user_id) AS is_owner
FROM group_members gm
JOIN groups g ON g.id = gm.group_id
WHERE gm.group_id = ANY($1::BIGINT[])
ORDER BY
    is_owner DESC,
    gm.joined_at ASC
`

type GetBatchedGroupMembersRow struct {
	ID       int64
	GroupID  int64
	UserID   int64
	JoinedAt pgtype.Timestamptz
	IsAdmin  bool
	IsOwner  bool
}

func (q *Queries) GetBatchedGroupMembers(ctx context.Context, groupIds []int64) ([]GetBatchedGroupMembersRow, error) {
	rows, err := q.db.Query(ctx, GetBatchedGroupMembers, groupIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBatchedGroupMembersRow
	for rows.Next() {
		var i GetBatchedGroupMembersRow
		if err := rows.Scan(
			&i.ID,
			&i.GroupID,
			&i.UserID,
			&i.JoinedAt,
			&i.IsAdmin,
			&i.IsOwner,
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

const GetBatchedGroups = `-- name: GetBatchedGroups :many
SELECT id, name, image, description, created_by, created_at FROM groups WHERE id = ANY($1::BIGINT[])
`

func (q *Queries) GetBatchedGroups(ctx context.Context, groupIds []int64) ([]Group, error) {
	rows, err := q.db.Query(ctx, GetBatchedGroups, groupIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Group
	for rows.Next() {
		var i Group
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Image,
			&i.Description,
			&i.CreatedBy,
			&i.CreatedAt,
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

const GetGroupByID = `-- name: GetGroupByID :one
SELECT id, name, image, description, created_by, created_at FROM groups WHERE id = $1
`

func (q *Queries) GetGroupByID(ctx context.Context, groupID int64) (Group, error) {
	row := q.db.QueryRow(ctx, GetGroupByID, groupID)
	var i Group
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Image,
		&i.Description,
		&i.CreatedBy,
		&i.CreatedAt,
	)
	return i, err
}

const GetGroupMembers = `-- name: GetGroupMembers :many
SELECT id, group_id, user_id, joined_at, is_admin FROM group_members WHERE group_id = $1
`

func (q *Queries) GetGroupMembers(ctx context.Context, groupID int64) ([]GroupMember, error) {
	rows, err := q.db.Query(ctx, GetGroupMembers, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GroupMember
	for rows.Next() {
		var i GroupMember
		if err := rows.Scan(
			&i.ID,
			&i.GroupID,
			&i.UserID,
			&i.JoinedAt,
			&i.IsAdmin,
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