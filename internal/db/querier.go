// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckMessagesHasNextPage(ctx context.Context, arg CheckMessagesHasNextPageParams) (bool, error)
	CheckMessagesHasPreviousPage(ctx context.Context, arg CheckMessagesHasPreviousPageParams) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	DeleteEmailVerificationToken(ctx context.Context, tokenID int64) error
	DeletePermission(ctx context.Context, name string) error
	DeleteRefreshToken(ctx context.Context, tokenID uuid.UUID) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error
	DeleteRole(ctx context.Context, name string) error
	DeleteRolePermissions(ctx context.Context, roleName string) error
	GetBatchedGroupMembers(ctx context.Context, groupIds []int64) ([]GetBatchedGroupMembersRow, error)
	GetBatchedGroups(ctx context.Context, groupIds []int64) ([]Group, error)
	GetBatchedMessages(ctx context.Context, messageIds []int64) ([]Message, error)
	GetBatchedUsers(ctx context.Context, userIds []int64) ([]GetBatchedUsersRow, error)
	GetChats(ctx context.Context, userID int64) ([]GetChatsRow, error)
	GetEmailVerificationToken(ctx context.Context, token string) (GetEmailVerificationTokenRow, error)
	GetGroupByID(ctx context.Context, groupID int64) (Group, error)
	GetGroupMembers(ctx context.Context, groupID int64) ([]GroupMember, error)
	GetMessages(ctx context.Context, arg GetMessagesParams) ([]Message, error)
	GetPermissions(ctx context.Context) ([]Permission, error)
	GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (RefreshToken, error)
	GetRoles(ctx context.Context) ([]Role, error)
	GetUser(ctx context.Context, userID int64) (GetUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserRoles(ctx context.Context, userID int64) ([]string, error)
	InsertEmailVerificationToken(ctx context.Context, arg InsertEmailVerificationTokenParams) (EmailVerificationToken, error)
	InsertMessage(ctx context.Context, arg InsertMessageParams) (Message, error)
	InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) error
	InsertRolePermission(ctx context.Context, arg InsertRolePermissionParams) error
	InsertUser(ctx context.Context, arg InsertUserParams) (User, error)
	InsertUserRole(ctx context.Context, arg InsertUserRoleParams) error
	UpdateUser(ctx context.Context, arg UpdateUserParams) error
	UpdateUserEmailVerifiedAt(ctx context.Context, arg UpdateUserEmailVerifiedAtParams) error
	UpdateUserOnlineStatus(ctx context.Context, arg UpdateUserOnlineStatusParams) error
	UpsertPermission(ctx context.Context, arg UpsertPermissionParams) error
	UpsertRole(ctx context.Context, arg UpsertRoleParams) error
}

var _ Querier = (*Queries)(nil)