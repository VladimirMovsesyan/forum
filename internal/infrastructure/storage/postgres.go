package storage

import (
	"context"
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"github.com/VladimirMovsesyan/forum/internal/domain/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
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
    allow_comments BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

	commentsTable = `CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INT REFERENCES posts(id) ON DELETE CASCADE,
    parent_id INT REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT CHECK (LENGTH(content) <= 2000),
    author TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
				VALUES ($1, $2, $3, $4) RETURNING id, created_at;`

	row := p.conn.QueryRow(ctx, query, post.Title, post.Content, post.Author, post.AllowComments)

	var id int

	var createdAt time.Time

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return &model.Post{}, err
	}

	return &model.Post{
		ID:            int32(id),
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		AllowComments: post.AllowComments,
		CreatedAt:     createdAt.String(),
		Comments:      make([]*model.Comment, 0),
	}, nil
}

func (p *postgresStorage) Post(ctx context.Context, id int) (*model.Post, error) {
	query := `SELECT * FROM posts WHERE id = $1;`

	row := p.conn.QueryRow(ctx, query, id)

	post := &model.Post{}

	var createdAt time.Time

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &createdAt)
	if err != nil {
		return nil, err
	}

	post.CreatedAt = createdAt.String()

	flatComments, err := p.Comments(ctx, id)
	if err != nil {
		return nil, err
	}

	commentMap := utils.BuildCommentTree(flatComments)

	var rootComments []*model.Comment
	for _, comment := range flatComments {
		if comment.ParentID == nil {
			rootComments = append(rootComments, commentMap[comment.ID])
		}
	}

	post.Comments = rootComments

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
		var createdAt time.Time

		post := &model.Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &createdAt)
		if err != nil {
			log.Println(err)
		}

		post.CreatedAt = createdAt.String()

		post.Comments, err = p.Comments(ctx, int(post.ID))

		posts = append(posts, post)
	}

	return posts, nil
}

func (p *postgresStorage) CreateComment(ctx context.Context, comment model.Comment) (*model.Comment, error) {
	query := `INSERT INTO comments (post_id, parent_id, content, author)
				VALUES ($1, $2, $3, $4)
				RETURNING id, created_at;`
	row := p.conn.QueryRow(ctx, query, comment.PostID, comment.ParentID, comment.Content, comment.Author)

	var id int

	var createdAt time.Time

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return &model.Comment{}, err
	}

	return &model.Comment{
		ID:        int32(id),
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: createdAt.String(),
	}, nil
}

func (p *postgresStorage) Comments(ctx context.Context, postID int) ([]*model.Comment, error) {
	query := `WITH RECURSIVE comment_tree AS (
				SELECT
					id,
					post_id,
					parent_id,
					author,
					content,
					created_at,
					id AS root_id
				FROM comments
				WHERE post_id = $1 AND parent_id IS NULL
			
				UNION ALL
			
				SELECT
					c.id,
					c.post_id,
					c.parent_id,
					c.author,
					c.content,
					c.created_at,
					ct.root_id
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_id = ct.id
			)
			SELECT id, post_id, parent_id, author, content, created_at FROM comment_tree
			ORDER BY root_id, created_at;`

	rows, err := p.conn.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	comments := make([]*model.Comment, 0)

	for rows.Next() {
		var createdAt time.Time

		comment := &model.Comment{}
		err = rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author, &createdAt)
		if err != nil {
			log.Println(err)
		}

		comment.CreatedAt = createdAt.String()

		comments = append(comments, comment)
	}

	return comments, nil
}
