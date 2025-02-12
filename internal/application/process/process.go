package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/VladimirMovsesyan/forum/internal/domain/pubsub"
	"github.com/VladimirMovsesyan/forum/internal/infrastructure/graphql"
	"github.com/VladimirMovsesyan/forum/internal/infrastructure/storage"
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

type repository interface {
	CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
	Post(ctx context.Context, id int) (*model.Post, error)
	Posts(ctx context.Context) ([]*model.Post, error)

	CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	Comments(ctx context.Context, postID int) ([]*model.Comment, error)
}

func (p *Process) Run() error {
	var s repository

	s = storage.NewInMemory()

	if p.dbDSN != "" {
		conn, err := pgxpool.New(context.Background(), p.dbDSN)
		if err != nil {
			return err
		}

		s, err = storage.NewPostgres(conn)
		if err != nil {
			return err
		}
	}

	ps := pubsub.NewPubSub()

	resolver := graphql.NewResolver(s, ps)

	srv := http.Server{
		Addr:              fmt.Sprintf(":%s", p.port),
		ReadHeaderTimeout: time.Second * 15,
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

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-signals
		log.Println("Got signal:", sig)

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	log.Println("Server gracefully stopped")

	return nil
}
