package dtloader

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
)

type MessageLoader = *dataloader.Loader[int64, model.Message]

func newMessageLoader(d db.DBQ) MessageLoader {
	cache := &dataloader.NoCache[int64, model.Message]{}

	loader := dataloader.NewBatchedLoader(func(ctx context.Context, ids []int64) []*dataloader.Result[model.Message] {
		results := make([]*dataloader.Result[model.Message], len(ids))

		res, err := d.GetBatchedMessages(ctx, ids)
		if err != nil {
			for idx := range ids {
				results[idx] = &dataloader.Result[model.Message]{
					Error: err,
				}
			}
			return results
		}

		msgMap := make(map[int64]*dataloader.Result[model.Message], len(ids))

		for _, m := range res {
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
				msgMap[m.ID] = &dataloader.Result[model.Message]{
					Error: err,
				}
				continue
			}

			msgMap[m.ID] = &dataloader.Result[model.Message]{
				Data: msg,
			}
		}

		for idx, id := range ids {
			results[idx] = msgMap[id]
		}

		return results
	}, dataloader.WithCache(cache))

	return loader
}
