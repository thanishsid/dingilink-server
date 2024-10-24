package dtloader

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
)

type GroupMembersLoader = *dataloader.Loader[int64, []*model.GroupMember]

func newGroupMembersLoader(d db.DBQ) GroupMembersLoader {
	cache := &dataloader.NoCache[int64, []*model.GroupMember]{}

	loader := dataloader.NewBatchedLoader(func(ctx context.Context, ids []int64) []*dataloader.Result[[]*model.GroupMember] {
		results := make([]*dataloader.Result[[]*model.GroupMember], len(ids))

		res, err := d.GetBatchedGroupMembers(ctx, ids)
		if err != nil {
			for idx := range ids {
				results[idx] = &dataloader.Result[[]*model.GroupMember]{
					Error: err,
				}
			}
			return results
		}

		groupsMap := make(map[int64][]*model.GroupMember, len(ids))

		for _, gm := range res {
			groupsMap[gm.GroupID] = append(groupsMap[gm.GroupID], &model.GroupMember{
				ID:       gm.ID,
				UserID:   gm.UserID,
				IsAdmin:  gm.IsAdmin,
				IsOwner:  gm.IsOwner,
				JoinedAt: gm.JoinedAt,
			})
		}

		for idx, id := range ids {
			results[idx] = &dataloader.Result[[]*model.GroupMember]{
				Data: groupsMap[id],
			}
		}

		return results
	}, dataloader.WithCache(cache))

	return loader
}
