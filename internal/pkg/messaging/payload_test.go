package messaging

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/model"
)

func TestSerializeAndDeserializePayload(t *testing.T) {
	RegisterType[model.MessageEvent]()

	var tp model.Message = model.TextMessage{
		ID:          3,
		SenderID:    5,
		RecipientID: null.IntFrom(3).Ptr(),
		TextContent: null.StringFrom("hellooo !!!").Ptr(),
		SentAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	me := model.MessageEvent{
		Type:      model.MessageEventTypeNew,
		MessageID: tp.GetID(),
	}

	b, err := SerializePayload(me)
	if err != nil {
		t.Fatal(err)
	}

	p, err := DeserializePayload(b)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.TypeOf(me) != reflect.TypeOf(p) {
		t.Errorf("type %T is not equal to %T", tp, p)
	}

	t.Logf("the original type is %T and the result type is: %T", me, p)
}
