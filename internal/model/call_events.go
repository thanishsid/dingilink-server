package model

type CallEventType string

const (
	CallEventTypeIncoming             = "incoming"
	CallEventTypeDeclined             = "declined"
	CallEventTypeIgnored              = "ignored"
	CallEventTypeTerminated           = "terminated"
	CallEventTypeIceCandidatesUpdated = "ice_updated"
)

type CallEvent struct {
	Type    CallEventType `json:"type"`
	CallID  int64         `json:"callId"`
	ActorID *int64        `json:"actorId"`
}
