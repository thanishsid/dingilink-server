package graphql

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/thanishsid/tokenizer"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/types/ctxt"
)

func webSocketInit(tc tokenizer.Config, d db.DBQ) transport.WebsocketInitFunc {
	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
		info := security.UserInfo{
			Authenticated: false,
		}

		authHeader := initPayload.Authorization()
		if authHeader == "" {
			return nil, &transport.InitPayload{}, fmt.Errorf("authorization header missing")
		}

		// Expecting format: "Bearer <token>"
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			return nil, &transport.InitPayload{}, fmt.Errorf("invalid Authorization format")
		}

		token := splitToken[1]

		if token != "" {
			var claims jwt.RegisteredClaims

			if err := tokenizer.ParseToken(ctx, tc, token, &claims); err != nil {
				return nil, &transport.InitPayload{}, fmt.Errorf("invalid token")
			}

			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				return nil, &transport.InitPayload{}, fmt.Errorf("invalid token")
			}

			tokenID, err := uuid.Parse(claims.ID)
			if err != nil {
				return nil, &transport.InitPayload{}, fmt.Errorf("invalid token")
			}

			userRoleNames, err := d.GetUserRoles(ctx, userID)
			if err != nil {
				return nil, &transport.InitPayload{}, fmt.Errorf("failed to get user roles")
			}

			roles := make([]security.Role, len(userRoleNames))

			for idx, r := range userRoleNames {
				roles[idx] = security.GetRoleByName(r)
			}

			info.Authenticated = true
			info.User = security.ContextUser{
				ID:      userID,
				Roles:   roles,
				TokenID: tokenID,
			}

			if err := d.UpdateUserOnlineStatus(ctx, db.UpdateUserOnlineStatusParams{
				Online: true,
				UserID: userID,
			}); err != nil {
				log.Println("failed to set user as online")
			}
		}

		ctxWithInfo := context.WithValue(ctx, ctxt.USER_INFO_CTX_KEY, info)

		return ctxWithInfo, &transport.InitPayload{}, nil
	}
}

func websocketClose(d db.DBQ) transport.WebsocketCloseFunc {
	return func(ctx context.Context, closeCode int) {
		userInfo, err := security.GetUserInfo(ctx)
		if err != nil {
			return
		}

		if userInfo.Authenticated {
			if err := d.UpdateUserOnlineStatus(ctx, db.UpdateUserOnlineStatusParams{
				Online: false,
				UserID: userInfo.User.ID,
			}); err != nil {
				log.Println("failed to set user as offline")
			}
		}
	}
}
