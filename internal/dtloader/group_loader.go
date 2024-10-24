package dtloader

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
)

type GroupLoader = *dataloader.Loader[int64, *model.Group]

func newGroupLoader(d db.DBQ) GroupLoader {
	cache := &dataloader.NoCache[int64, *model.Group]{}

	loader := dataloader.NewBatchedLoader(func(ctx context.Context, ids []int64) []*dataloader.Result[*model.Group] {
		results := make([]*dataloader.Result[*model.Group], len(ids))

		res, err := d.GetBatchedGroups(ctx, ids)
		if err != nil {
			for idx := range ids {
				results[idx] = &dataloader.Result[*model.Group]{
					Error: err,
				}
			}
			return results
		}

		groupsMap := make(map[int64]*dataloader.Result[*model.Group], len(ids))

		for _, g := range res {
			groupsMap[g.ID] = &dataloader.Result[*model.Group]{
				Data: &model.Group{
					ID:          g.ID,
					Name:        g.Name,
					Description: g.Description,
					Image:       g.Image,
					CreatedBy:   g.CreatedBy,
				},
			}
		}

		for idx, id := range ids {
			results[idx] = groupsMap[id]
		}

		return results
	}, dataloader.WithCache(cache))

	return loader
}
