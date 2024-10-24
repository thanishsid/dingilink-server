package dtloader

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
)

type UserLoader = *dataloader.Loader[int64, *model.User]

func newUserLoader(d db.DBQ) UserLoader {
	cache := &dataloader.NoCache[int64, *model.User]{}

	loader := dataloader.NewBatchedLoader(func(ctx context.Context, ids []int64) []*dataloader.Result[*model.User] {
		results := make([]*dataloader.Result[*model.User], len(ids))

		res, err := d.GetBatchedUsers(ctx, ids)
		if err != nil {
			for idx := range ids {
				results[idx] = &dataloader.Result[*model.User]{
					Error: err,
				}
			}
			return results
		}

		usersMap := make(map[int64]*dataloader.Result[*model.User], len(ids))

		for _, u := range res {
			usersMap[u.ID] = &dataloader.Result[*model.User]{
				Data: &model.User{
					ID:          u.ID,
					Username:    u.Username,
					Email:       u.Email,
					Name:        u.Name,
					Bio:         u.Bio,
					Image:       u.Image,
					Online:      u.Online,
					FriendCount: u.FriendCount,
				},
			}
		}

		for idx, id := range ids {
			results[idx] = usersMap[id]
		}

		return results
	}, dataloader.WithCache(cache))

	return loader
}
