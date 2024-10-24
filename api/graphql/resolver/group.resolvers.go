package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/thanishsid/dingilink-server/api/graphql/generated"
	"github.com/thanishsid/dingilink-server/internal/model"
)

// Members is the resolver for the members field.
func (r *groupResolver) Members(ctx context.Context, obj *model.Group) ([]*model.GroupMember, error) {
	return r.Dataloader.GetGroupMembers(ctx, obj.ID)
}

// User is the resolver for the user field.
func (r *groupMemberResolver) User(ctx context.Context, obj *model.GroupMember) (*model.User, error) {
	return r.Dataloader.GetUser(ctx, obj.UserID)
}

// Group returns generated.GroupResolver implementation.
func (r *Resolver) Group() generated.GroupResolver { return &groupResolver{r} }

// GroupMember returns generated.GroupMemberResolver implementation.
func (r *Resolver) GroupMember() generated.GroupMemberResolver { return &groupMemberResolver{r} }

type groupResolver struct{ *Resolver }
type groupMemberResolver struct{ *Resolver }