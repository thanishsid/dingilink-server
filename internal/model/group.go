package model

import "github.com/jackc/pgx/v5/pgtype"

type Group struct {
	ID          int64
	Name        string
	Description *string
	Image       *string
	CreatedBy   int64
}

type GroupMember struct {
	ID       int64
	UserID   int64
	IsAdmin  bool
	IsOwner  bool
	JoinedAt pgtype.Timestamptz
}
