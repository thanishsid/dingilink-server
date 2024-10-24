package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	vd "github.com/go-ozzo/ozzo-validation/v4"
	"gopkg.in/guregu/null.v4"
	"gopkg.in/typ.v4/slices"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
	"github.com/thanishsid/dingilink-server/internal/pkg/messaging"
	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/types"
	"github.com/thanishsid/dingilink-server/internal/types/apperror"
)

type MessageService struct {
	DB db.DBQ
	CH *messaging.ChannelManager[*model.MessageEvent]
}

func getMessageChannelID(userID int64) string {
	return fmt.Sprintf("user_%d.message_events", userID)
}

// Subscribe to message events
func (s *MessageService) SubscribeToMessageEvents(ctx context.Context) (<-chan *model.MessageEvent, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	channelID := getMessageChannelID(userInfo.User.ID)

	ch, err := s.CH.Subscribe(channelID)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		s.CH.Unsubscribe(channelID)
	}()

	return ch, nil
}

// Get all user chats
func (s *MessageService) GetChats(ctx context.Context) ([]model.ChatPreview, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	chatsResult, err := s.DB.GetChats(ctx, userInfo.User.ID)
	if err != nil {
		return nil, err
	}

	chats := make([]model.ChatPreview, len(chatsResult))

	for idx, c := range chatsResult {
		var chat model.ChatPreview

		if c.IsGroupChat {
			chat = model.GroupChatPreview{
				GroupID:            null.IntFromPtr(c.ChatID).ValueOrZero(),
				LastMessageID:      c.LastMessageID,
				UnreadMessageCount: c.UnreadMessageCount,
			}
		} else {
			chat = model.DirectChatPreview{
				UserID:             null.IntFromPtr(c.ChatID).ValueOrZero(),
				LastMessageID:      c.LastMessageID,
				UnreadMessageCount: c.UnreadMessageCount,
			}
		}

		chats[idx] = chat
	}

	return chats, nil
}

// Get a chat
func (s *MessageService) GetChat(ctx context.Context, chatID string) (model.Chat, error) {
	_, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	splitId := strings.Split(chatID, "_")
	if len(splitId) != 2 {
		return nil, fmt.Errorf("invalid id")
	}

	idType := splitId[0]

	id := parseID(splitId[1])

	switch idType {
	case "group":
		return &model.GroupChat{
			GroupID: id,
		}, nil
	case "direct":
		return &model.DirectChat{
			UserID: id,
		}, nil
	}

	return nil, fmt.Errorf("invalid id")
}

type GetMessagesInput struct {
	Last   *int64
	Before *string
}

func (s *MessageService) GetMessages(ctx context.Context, chatID string, input GetMessagesInput) (*model.MessageConnection, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	splitId := strings.Split(chatID, "_")
	if len(splitId) != 2 {
		return nil, fmt.Errorf("invalid id")
	}

	idType := splitId[0]

	id := parseID(splitId[1])

	var cursorID *int64
	var limit int64 = 30

	if input.Before != nil {
		cID, err := strconv.ParseInt(*input.Before, 10, 64)
		if err != nil {
			return nil, err
		}

		cursorID = &cID
	}

	if input.Last != nil {
		limit = *input.Last
	}

	messagesResult, err := s.DB.GetMessages(ctx, db.GetMessagesParams{
		TargetUserID:  null.NewInt(id, idType == "direct").Ptr(),
		TargetGroupID: null.NewInt(id, idType == "group").Ptr(),
		CurrentUserID: userInfo.User.ID,
		CursorID:      cursorID,
		ResultLimit:   limit,
	})
	if err != nil {
		return nil, err
	}

	edges := make([]model.MessageEdge, len(messagesResult))

	for idx, m := range messagesResult {
		msg, err := model.MessageBuilder{
			ID:                m.ID,
			MessageType:       m.MessageType,
			SenderID:          m.SenderID,
			RecipientID:       m.RecipientID,
			GroupID:           m.GroupID,
			TextContent:       m.TextContent,
			Media:             m.Media,
			Location:          m.Location,
			ReplyForMessageID: m.ReplyForMessageID,
			SentAt:            m.SentAt,
			DeletedAt:         m.DeletedAt,
			DeletedBy:         m.DeletedBy,
		}.Build()
		if err != nil {
			return nil, err
		}

		edge := model.MessageEdge{
			Node:   msg,
			Cursor: fmt.Sprint(msg.GetID()),
		}

		edges[idx] = edge
	}

	connection := model.MessageConnection{
		Edges:    edges,
		PageInfo: model.PageInfo{},
	}

	if len(edges) > 0 {
		hasNextPage, err := s.DB.CheckMessagesHasNextPage(ctx, db.CheckMessagesHasNextPageParams{
			TargetUserID:  null.NewInt(id, idType == "direct").Ptr(),
			TargetGroupID: null.NewInt(id, idType == "group").Ptr(),
			CurrentUserID: userInfo.User.ID,
			CursorID:      edges[len(edges)-1].Node.GetID(),
		})
		if err != nil {
			return nil, err
		}

		hasPreviousPage, err := s.DB.CheckMessagesHasPreviousPage(ctx, db.CheckMessagesHasPreviousPageParams{
			TargetUserID:  null.NewInt(id, idType == "direct").Ptr(),
			TargetGroupID: null.NewInt(id, idType == "group").Ptr(),
			CurrentUserID: userInfo.User.ID,
			CursorID:      edges[0].Node.GetID(),
		})
		if err != nil {
			return nil, err
		}

		connection.PageInfo.HasNextPage = hasNextPage
		connection.PageInfo.HasPreviousPage = hasPreviousPage
		connection.PageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &connection, nil
}

type SendMessageInput struct {
	GroupID           *int64            `json:"groupId"`
	UserID            *int64            `json:"userId"`
	Type              model.MessageType `json:"type"`
	Text              *string           `json:"text"`
	Media             *string           `json:"media"`
	Location          *types.LatLng     `json:"location"`
	ReplyForMessageID *int64            `json:"replyForMessageId"`
}

func (i SendMessageInput) Validate() error {

	allowedMessageTypes := slices.Map([]model.MessageType{
		model.MessageTypeText,
		model.MessageTypeImage,
		model.MessageTypeAudio,
		model.MessageTypeVideo,
		model.MessageTypeDocument,
		model.MessageTypeLocation,
	}, func(mt model.MessageType) any {
		return mt
	})

	return vd.ValidateStruct(&i,
		vd.Field(&i.GroupID,
			vd.Required.When(vd.IsEmpty(i.UserID)).Error(apperror.INPUT_REQUIRED),
			vd.Nil.When(!vd.IsEmpty(i.UserID)).Error(apperror.INPUT_NOT_REQUIRED),
		),
		vd.Field(&i.UserID,
			vd.Required.When(vd.IsEmpty(i.GroupID)).Error(apperror.INPUT_REQUIRED),
			vd.Nil.When(!vd.IsEmpty(i.GroupID)).Error(apperror.INPUT_NOT_REQUIRED),
		),
		vd.Field(&i.Type,
			vd.In(
				allowedMessageTypes...,
			),
		),
		vd.Field(&i.Text, vd.Required.When(i.Type == model.MessageTypeText).Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.Media, vd.Required.When(slices.Contains([]model.MessageType{
			model.MessageTypeImage,
			model.MessageTypeAudio,
			model.MessageTypeVideo,
			model.MessageTypeDocument,
		}, i.Type)).Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.Location, vd.Required.When(i.Type == model.MessageTypeLocation).Error(apperror.INPUT_REQUIRED)),
	)
}

// Send a new message.
func (s *MessageService) SendMessage(ctx context.Context, input SendMessageInput) (model.Message, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	m, err := s.DB.InsertMessage(ctx, db.InsertMessageParams{
		SenderID:          userInfo.User.ID,
		RecipientID:       input.UserID,
		GroupID:           input.GroupID,
		MessageType:       input.Type.String(),
		TextContent:       input.Text,
		Media:             input.Media,
		Location:          types.LatLngToPoint(input.Location),
		ReplyForMessageID: input.ReplyForMessageID,
	})
	if err != nil {
		return nil, err
	}

	msgb := model.MessageBuilder{
		ID:                m.ID,
		SenderID:          m.SenderID,
		RecipientID:       m.RecipientID,
		GroupID:           m.GroupID,
		MessageType:       m.MessageType,
		TextContent:       m.TextContent,
		Media:             m.Media,
		Location:          m.Location,
		ReplyForMessageID: m.ReplyForMessageID,
		SentAt:            m.SentAt,
		DeletedAt:         m.DeletedAt,
		DeletedBy:         m.DeletedBy,
	}

	isDirect := m.RecipientID != nil
	isGroup := m.GroupID != nil

	messageEvent := model.MessageEvent{
		Type:      model.MessageEventTypeNew,
		MessageID: m.ID,
	}

	go func() {
		// If direct message send an event through the channel manager to the recipient
		if isDirect {
			if err := s.CH.SendPayload(getMessageChannelID(null.IntFromPtr(m.RecipientID).ValueOrZero()), &messageEvent); err != nil {
				log.Printf("failed to send direct message event via channel manager: %v", err)
			}
		}

		// If group message send an event through the channel manager to all group members
		if isGroup {
			members, err := s.DB.GetGroupMembers(ctx, *m.GroupID)
			if err != nil {
				return
			}

			for _, mb := range members {
				if err := s.CH.SendPayload(getMessageChannelID(mb.UserID), &messageEvent); err != nil {
					log.Printf("failed to send group message event via channel manager: %v", err)
				}
			}
		}
	}()

	msg, err := msgb.Build()
	if err != nil {
		return nil, err
	}

	return msg, nil
}
