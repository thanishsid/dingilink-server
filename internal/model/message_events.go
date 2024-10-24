package model

type MessageEventType string

const (
	MessageEventTypeNew     = "new"
	MessageEventTypeDeleted = "deleted"
	MessageEventTypeEdited  = "edited"
)

type MessageEvent struct {
	Type      MessageEventType `json:"type"`
	MessageID int64            `json:"messageId"`
}

func (me MessageEvent) ID() int64 {
	return me.MessageID
}
