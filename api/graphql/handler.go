package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/thanishsid/tokenizer"

	"github.com/thanishsid/dingilink-server/api/graphql/generated"
	"github.com/thanishsid/dingilink-server/api/graphql/resolver"
	"github.com/thanishsid/dingilink-server/internal/db"
)

func NewHandler(r *resolver.Resolver, tc tokenizer.Config, d db.DBQ) *handler.Server {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: r,
	}))

	h.AddTransport(transport.Websocket{
		InitFunc:  webSocketInit(tc, d),
		CloseFunc: websocketClose(d),
	})

	// Set panic recovery function.
	h.SetRecoverFunc(recoverFunc)

	// Set Error presenter.
	h.SetErrorPresenter(ErrPresenter)

	return h
}
