package api

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/thanishsid/tokenizer"

	"github.com/thanishsid/dingilink-server/api/graphql"
	"github.com/thanishsid/dingilink-server/api/graphql/resolver"
	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/dtloader"
	"github.com/thanishsid/dingilink-server/internal/services"
)

type HandlerConfig struct {
	UploadService  *services.UploadService
	UserService    *services.UserService
	MessageService *services.MessageService

	PG db.DBQ
	TC tokenizer.Config
}

func NewHandler(hc *HandlerConfig) http.Handler {

	dataloader := dtloader.NewDataloader(hc.PG)

	gqlHandler := graphql.NewHandler(&resolver.Resolver{
		UserService:    hc.UserService,
		MessageService: hc.MessageService,
		Dataloader:     dataloader,
	}, hc.TC, hc.PG)

	r := chi.NewRouter()

	// Set CORS options
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"X-Authorization",
			"X-Real-Ip",
			"x-client-location",
		},
		MaxAge: 300,
	}))

	// RealIP middleware looks for the client ip address in headers and updates the http.Request.RemoteAddr value.
	r.Use(chimiddleware.RealIP)

	r.Use(UserInfoMiddleware(hc.TC, hc.PG))

	// File uploads
	r.Post("/upload", CreateUploadHandler(hc.UploadService))

	// GraphQL Handler
	r.Mount("/", gqlHandler)

	// GraphQL Playground
	r.Mount("/graphiql", playground.Handler("GraphQL playground", "/graphql"))

	return r
}
