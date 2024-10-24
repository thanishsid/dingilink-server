package model

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/types"
)

type Message interface {
	IsMessage()
	GetID() int64
}

type MessageType string

func (mt MessageType) String() string {
	return string(mt)
}

const (
	MessageTypeText     = "text"
	MessageTypeImage    = "image"
	MessageTypeAudio    = "audio"
	MessageTypeVideo    = "video"
	MessageTypeDocument = "document"
	MessageTypeLocation = "location"
)

type GenericMessage[T any] struct {
	ID          int64
	SenderID    int64
	RecipientID *int64
	GroupID     *int64
	TextContent *string
	Payload     T
	ParentID    *int64
	SentAt      pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
	DeletedBy   *int64
}

func getChatID(ctx context.Context, senderID int64,
	recipientID *int64,
	groupID *int64) (string, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return "", err
	}

	if groupID != nil {
		return fmt.Sprintf("group_%d", *groupID), nil
	}

	if recipientID != nil {
		if *recipientID == userInfo.User.ID {
			return fmt.Sprintf("direct_%d", senderID), nil
		} else {
			return fmt.Sprintf("direct_%d", *recipientID), nil
		}
	}

	return "", fmt.Errorf("failed to determine chat id")
}

type LocationMessagePayload struct {
	Location types.Point
}

// Text Message
type TextMessage GenericMessage[any]

func (TextMessage) IsMessage()     {}
func (m TextMessage) GetID() int64 { return m.ID }
func (m TextMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

// Image Message
type ImageMessage GenericMessage[string]

func (ImageMessage) IsMessage()     {}
func (m ImageMessage) GetID() int64 { return m.ID }
func (m ImageMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

// Audio Message
type AudioMessage GenericMessage[string]

func (AudioMessage) IsMessage()     {}
func (m AudioMessage) GetID() int64 { return m.ID }
func (m AudioMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

// Video Message
type VideoMessage GenericMessage[string]

func (VideoMessage) IsMessage()     {}
func (m VideoMessage) GetID() int64 { return m.ID }
func (m VideoMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}
func (m VideoMessage) Thumbnail() (*string, error) {
	metadata, err := DecodeObjectMetadata(m.Payload)
	if err != nil {
		return nil, err
	}

	return metadata.Thumbnail, nil
}

// Document Message
type DocumentMessage GenericMessage[string]

func (DocumentMessage) IsMessage()     {}
func (m DocumentMessage) GetID() int64 { return m.ID }
func (m DocumentMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

// Location Message
type LocationMessage GenericMessage[LocationMessagePayload]

func (LocationMessage) IsMessage()     {}
func (m LocationMessage) GetID() int64 { return m.ID }
func (m LocationMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

// Deleted Message
type DeletedMessage GenericMessage[any]

func (DeletedMessage) IsMessage()     {}
func (m DeletedMessage) GetID() int64 { return m.ID }
func (m DeletedMessage) ChatID(ctx context.Context) (string, error) {
	return getChatID(ctx, m.SenderID, m.RecipientID, m.GroupID)
}

//----- MESSAGE BUILDER ----->

type MessageBuilder struct {
	ID                int64              `json:"id"`
	SenderID          int64              `json:"senderId"`
	RecipientID       *int64             `json:"RecipientId"`
	GroupID           *int64             `json:"groupId"`
	MessageType       string             `json:"messageType"`
	TextContent       *string            `json:"textContent"`
	Media             *string            `json:"media"`
	Location          types.Point        `json:"location"`
	ReplyForMessageID *int64             `json:"replyForMessageId"`
	SentAt            pgtype.Timestamptz `json:"sentAt"`
	DeletedAt         pgtype.Timestamptz `json:"deletedAt"`
	DeletedBy         *int64             `json:"deletedBy"`
}

func (m MessageBuilder) Build() (Message, error) {
	var msg Message

	if m.DeletedAt.Valid {
		return DeletedMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
			DeletedAt:   m.DeletedAt,
			DeletedBy:   m.DeletedBy,
		}, nil
	}

	switch m.MessageType {
	case "text":
		msg = TextMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
		}
	case "image":
		msg = ImageMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
			Payload:     null.StringFromPtr(m.Media).ValueOrZero(),
		}
	case "audio":
		msg = AudioMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
			Payload:     null.StringFromPtr(m.Media).ValueOrZero(),
		}
	case "video":
		msg = VideoMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
			Payload:     null.StringFromPtr(m.Media).ValueOrZero(),
		}
	case "location":
		msg = LocationMessage{
			ID:          m.ID,
			SenderID:    m.SenderID,
			RecipientID: m.RecipientID,
			GroupID:     m.GroupID,
			TextContent: m.TextContent,
			ParentID:    m.ReplyForMessageID,
			SentAt:      m.SentAt,
			Payload: LocationMessagePayload{
				Location: m.Location,
			},
		}
	default:
		return nil, fmt.Errorf("invalid message type")
	}

	return msg, nil
}

// Message Connection

type MessageEdge = Edge[Message]
type MessageConnection Connection[Message]
