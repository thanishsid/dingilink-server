package model

type User struct {
	ID          int64
	Username    string
	Email       string
	Name        string
	Bio         *string
	Image       *string
	Online      bool
	FriendCount int64
}

type Role struct {
	ID          int64
	Name        string
	Description string
}
