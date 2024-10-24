package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/thanishsid/tokenizer"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/types"
	"github.com/thanishsid/dingilink-server/internal/types/ctxt"
)

func UserInfoMiddleware(tc tokenizer.Config, d db.DBQ) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			info := security.UserInfo{
				Authenticated: false,
			}

			var authHeader string

			for _, key := range []string{"authorization", "x-authorization"} {
				authHeader = r.Header.Get(key)
				if authHeader != "" {
					break
				}
			}

			if authHeader != "" {

				// Expecting format: "Bearer <token>"
				splitToken := strings.Split(authHeader, "Bearer ")
				if len(splitToken) != 2 {
					http.Error(w, "invalid authorization format", http.StatusUnauthorized)
					return
				}

				token := splitToken[1]

				var claims jwt.RegisteredClaims

				if err := tokenizer.ParseToken(r.Context(), tc, token, &claims); err != nil {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				userID, err := strconv.ParseInt(claims.Subject, 10, 64)
				if err != nil {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				tokenID, err := uuid.Parse(claims.ID)
				if err != nil {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				userRoleNames, err := d.GetUserRoles(r.Context(), userID)
				if err != nil {
					http.Error(w, "failed to get user roles", http.StatusUnauthorized)
					return
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

			}

			clientLocation, err := getClientLocation(r)
			if err == nil {
				info.Location = clientLocation
			}

			ctxWithInfo := context.WithValue(r.Context(), ctxt.USER_INFO_CTX_KEY, info)

			next.ServeHTTP(w, r.WithContext(ctxWithInfo))
		})

	}
}

func getClientLocation(r *http.Request) (types.Point, error) {
	var point types.Point

	locationHeaderVal := r.Header.Get("x-client-location")

	if locationHeaderVal == "" {
		return point, errors.New("x-client-location header is empty")
	}

	coordinateStringParts := strings.Split(locationHeaderVal, ":")

	if len(coordinateStringParts) != 2 {
		return point, errors.New("invalid location format")
	}

	lat, err := strconv.ParseFloat(coordinateStringParts[0], 64)
	if err != nil {
		return point, errors.New("invalid latitude")
	}

	lng, err := strconv.ParseFloat(coordinateStringParts[1], 64)
	if err != nil {
		return point, errors.New("invalid longitude")
	}

	point = types.NewCoordinates(lat, lng)

	return point, nil
}
