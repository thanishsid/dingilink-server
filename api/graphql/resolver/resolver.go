package resolver

import (
	"github.com/thanishsid/dingilink-server/internal/dtloader"
	"github.com/thanishsid/dingilink-server/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// Services
	UserService    *services.UserService
	MessageService *services.MessageService

	// Dataloader
	Dataloader *dtloader.Dataloader
}
