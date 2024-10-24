package model

import (
	"fmt"
)

type Chat interface {
	IsChat()
	ID() string
}

type ChatPreview interface {
	IsChatPreview()
	ID() string
}

type DirectChatPreview struct {
	UserID             int64
	LastMessageID      int64
	UnreadMessageCount int64
}

func (DirectChatPreview) IsChatPreview() {}
func (c DirectChatPreview) ID() string {
	return fmt.Sprintf("direct_%d", c.UserID)
}

type GroupChatPreview struct {
	GroupID            int64
	LastMessageID      int64
	UnreadMessageCount int64
}

func (GroupChatPreview) IsChatPreview() {}
func (c GroupChatPreview) ID() string {
	return fmt.Sprintf("group_%d", c.GroupID)
}

type DirectChat struct {
	UserID int64
}

func (DirectChat) IsChat() {}
func (c DirectChat) ID() string {
	return fmt.Sprintf("direct_%d", c.UserID)
}

type GroupChat struct {
	GroupID int64
}

func (GroupChat) IsChat() {}
func (c GroupChat) ID() string {
	return fmt.Sprintf("group_%d", c.GroupID)
}
