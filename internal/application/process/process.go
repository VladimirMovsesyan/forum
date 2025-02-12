package process

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/VladimirMovsesyan/forum/internal/domain/pubsub"
	"github.com/VladimirMovsesyan/forum/internal/infrastructure/graphql"
	"github.com/VladimirMovsesyan/forum/internal/infrastructure/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
	"log"
	"net/http"
	"time"
)

type Process struct {
	port  string
	dbDSN string
}

func New(port, dbDSN string) *Process {
	return &Process{
		port:  port,
		dbDSN: dbDSN,
	}
}

func (p *Process) Run() error {
	conn, err := pgxpool.New(context.Background(), p.dbDSN)
	if err != nil {
		return err
	}

	s, err := storage.NewPostgres(conn)
	if err != nil {
		return err
	}

	ps := pubsub.NewPubSub()

	resolver := graphql.NewResolver(s, ps)

	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: &resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground\n", p.port)

	if err := http.ListenAndServe(":"+p.port, nil); errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
