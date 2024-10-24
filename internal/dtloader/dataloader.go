package dtloader

import (
	"context"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
)

// func NewMiddleware(dl *Dataloader) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ctx := context.WithValue(r.Context(), ctxt.DATALOADER_CTX_KEY, dl)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

// func DataloaderFor(ctx context.Context) *Dataloader {
// 	loader, ok := ctx.Value(ctxt.DATALOADER_CTX_KEY).(*Dataloader)
// 	if !ok {
// 		panic("dataloader not found in context")
// 	}

// 	return loader
// }

type Dataloader struct {
	user         UserLoader
	group        GroupLoader
	groupMembers GroupMembersLoader
	message      MessageLoader
}

func NewDataloader(d db.DBQ) *Dataloader {
	return &Dataloader{
		user:         newUserLoader(d),
		group:        newGroupLoader(d),
		groupMembers: newGroupMembersLoader(d),
		message:      newMessageLoader(d),
	}
}

// Get a user by id.
func (d *Dataloader) GetUser(ctx context.Context, userID int64) (*model.User, error) {
	return d.user.Load(ctx, userID)()
}

// Get a group by id.
func (d *Dataloader) GetGroup(ctx context.Context, groupID int64) (*model.Group, error) {
	return d.group.Load(ctx, groupID)()
}

// Get the members of a group by the group id
func (d *Dataloader) GetGroupMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error) {
	return d.groupMembers.Load(ctx, groupID)()
}

// Get a message by id.
func (d *Dataloader) GetMessage(ctx context.Context, messageID int64) (model.Message, error) {
	return d.message.Load(ctx, messageID)()
}
