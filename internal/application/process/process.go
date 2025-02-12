package process

import (
	"context"
	"errors"
	"fmt"
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
	"os"
	"os/signal"
	"syscall"
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

	srv := http.Server{
		Addr: fmt.Sprintf(":%s", p.port),
	}

	queryHandler := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: &resolver}))

	queryHandler.AddTransport(transport.Options{})
	queryHandler.AddTransport(transport.GET{})
	queryHandler.AddTransport(transport.POST{})
	queryHandler.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	queryHandler.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	queryHandler.Use(extension.Introspection{})
	queryHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", queryHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground\n", p.port)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	go func() {
		sig := <-signals
		log.Println("Got signal:", sig)
		err := srv.Shutdown(context.Background())
		if err != nil {
			log.Println(err)
		}
	}()

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	log.Println("Server gracefully stopped")

	return nil
}
