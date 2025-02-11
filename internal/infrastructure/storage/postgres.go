package storage

import (
	"context"
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strconv"
)

type repository interface {
	CreatePost(ctx context.Context, post model.Post) (*model.Post, error)
	Post(ctx context.Context, id int) (*model.Post, error)
	Posts(ctx context.Context) ([]*model.Post, error)

	CreateComment(ctx context.Context, comment model.Comment) (*model.Comment, error)
	Comments(ctx context.Context, postID int) ([]*model.Comment, error)
}

var _ repository = &postgresStorage{}

const (
	postsTable = `CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    allow_comments BOOLEAN DEFAULT TRUE
);`

	commentsTable = `CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INT REFERENCES posts(id),
    parent_id INT REFERENCES comments(id),
    content TEXT CHECK (LENGTH(content) <= 2000),
    author TEXT NOT NULL
);`
)

type postgresStorage struct {
	conn *pgxpool.Pool
}

func NewPostgres(conn *pgxpool.Pool) (*postgresStorage, error) {
	storage := &postgresStorage{
		conn: conn,
	}

	err := storage.ensureTablesExists()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (p *postgresStorage) ensureTablesExists() error {
	_, err := p.conn.Exec(context.Background(), postsTable)
	if err != nil {
		return err
	}

	_, err = p.conn.Exec(context.Background(), commentsTable)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresStorage) CreatePost(ctx context.Context, post model.Post) (*model.Post, error) {
	query := `INSERT INTO posts (title, content, author, allow_comments)
				VALUES ($1, $2, $3, $4) RETURNING id;`

	row := p.conn.QueryRow(ctx, query, post.Title, post.Content, post.Author, post.AllowComments)

	var id int

	err := row.Scan(&id)
	if err != nil {
		return &model.Post{}, err
	}

	return &model.Post{
		ID:            strconv.Itoa(id),
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		AllowComments: post.AllowComments,
		Comments:      make([]*model.Comment, 0),
	}, nil
}

func (p *postgresStorage) Post(ctx context.Context, id int) (*model.Post, error) {
	query := `SELECT * FROM posts WHERE id = $1;`

	row := p.conn.QueryRow(ctx, query, id)

	post := &model.Post{}

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments)
	if err != nil {
		return nil, err
	}

	post.Comments, err = p.Comments(ctx, id)

	return post, nil
}

func (p *postgresStorage) Posts(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT * FROM posts;`
	rows, err := p.conn.Query(ctx, query)
	if err != nil {
		return []*model.Post{}, err
	}

	posts := make([]*model.Post, 0)

	for rows.Next() {
		post := &model.Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments)
		if err != nil {
			log.Println(err)
		}

		id, err := strconv.Atoi(post.ID)
		if err != nil {
			return nil, err
		}

		post.Comments, err = p.Comments(ctx, id)

		posts = append(posts, post)
	}

	return posts, nil
}

func (p *postgresStorage) CreateComment(ctx context.Context, comment model.Comment) (*model.Comment, error) {
	query := `INSERT INTO comments (post_id, parent_id, content, author)
				VALUES ($1, $2, $3, $4)
				RETURNING id;`
	row := p.conn.QueryRow(ctx, query, comment.PostID, comment.ParentID, comment.Content, comment.Author)

	var id int

	err := row.Scan(&id)
	if err != nil {
		return &model.Comment{}, err
	}

	return &model.Comment{
		ID:       strconv.Itoa(id),
		PostID:   comment.PostID,
		ParentID: comment.ParentID,
		Content:  comment.Content,
		Author:   comment.Author,
	}, nil
}

func (p *postgresStorage) Comments(ctx context.Context, postID int) ([]*model.Comment, error) {
	query := `SELECT * FROM comments WHERE post_id = $1;`

	rows, err := p.conn.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	comments := make([]*model.Comment, 0)

	for rows.Next() {
		comment := &model.Comment{}
		err = rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author)
		if err != nil {
			log.Println(err)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
