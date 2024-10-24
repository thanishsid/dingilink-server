package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/thanishsid/dingilink-server/internal/model"
	"github.com/thanishsid/dingilink-server/internal/services"
)

// Register is the resolver for the register field.
func (r *mutationsResolver) Register(ctx context.Context, input services.RegistrationInput) (bool, error) {
	if err := r.UserService.Register(ctx, input); err != nil {
		return fail(err)
	}

	return success()
}

// VerifyEmail is the resolver for the verifyEmail field.
func (r *mutationsResolver) VerifyEmail(ctx context.Context, input services.EmailVerificationInput) (*model.TokenPair, error) {
	return r.UserService.VerifyEmail(ctx, input)
}

// ResendEmailVerification is the resolver for the resendEmailVerification field.
func (r *mutationsResolver) ResendEmailVerification(ctx context.Context, input services.ResendEmailVerificationInput) (bool, error) {
	if err := r.UserService.ResendEmailVerification(ctx, input); err != nil {
		return fail(err)
	}

	return success()
}

// Login is the resolver for the login field.
func (r *mutationsResolver) Login(ctx context.Context, input services.LoginInput) (*model.TokenPair, error) {
	return r.UserService.Login(ctx, input)
}

// RefreshTokens is the resolver for the refreshTokens field.
func (r *mutationsResolver) RefreshTokens(ctx context.Context, input services.RefreshTokensInput) (*model.TokenPair, error) {
	return r.UserService.RefreshTokens(ctx, input)
}

// Logout is the resolver for the logout field.
func (r *mutationsResolver) Logout(ctx context.Context) (bool, error) {
	if err := r.UserService.Logout(ctx); err != nil {
		return fail(err)
	}

	return success()
}

// LogoutFromAllDevices is the resolver for the logoutFromAllDevices field.
func (r *mutationsResolver) LogoutFromAllDevices(ctx context.Context) (bool, error) {
	if err := r.UserService.LogoutFromAllDevices(ctx); err != nil {
		return fail(err)
	}

	return success()
}

// UpdateCurrentUser is the resolver for the updateCurrentUser field.
func (r *mutationsResolver) UpdateCurrentUser(ctx context.Context, input services.UpdateCurrentUserInput) (*model.User, error) {
	return r.UserService.UpdateCurrentUser(ctx, input)
}

// CurrentUser is the resolver for the currentUser field.
func (r *queriesResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	return r.UserService.GetCurrentUser(ctx)
}
