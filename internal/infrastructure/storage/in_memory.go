package storage

import (
	"context"
	"fmt"
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"github.com/VladimirMovsesyan/forum/internal/domain/utils"
	"log"
	"sort"
	"time"
)

type repository interface {
	CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
	Post(ctx context.Context, id int) (*model.Post, error)
	Posts(ctx context.Context) ([]*model.Post, error)

	CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	Comments(ctx context.Context, postID int) ([]*model.Comment, error)
}

var _ repository = &InMemory{}

type InMemory struct {
	postIDs    int32
	posts      map[int32]*model.Post
	commentIDs int32
	comments   map[int32]*model.Comment
}

func NewInMemory() *InMemory {
	return &InMemory{
		posts:    make(map[int32]*model.Post),
		comments: make(map[int32]*model.Comment),
	}
}

func (mem *InMemory) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	mem.postIDs += 1
	mem.posts[mem.postIDs] = &model.Post{
		ID:            mem.postIDs,
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		AllowComments: post.AllowComments,
		CreatedAt:     time.Now().String(),
		Comments:      make([]*model.Comment, 0),
	}

	return mem.posts[mem.postIDs], nil
}

func (mem *InMemory) Post(ctx context.Context, id int) (*model.Post, error) {
	post, ok := mem.posts[int32(id)]
	if !ok {
		return nil, fmt.Errorf("post not found with id %d", id)
	}

	flatComments, err := mem.Comments(ctx, id)
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

func (mem *InMemory) Posts(ctx context.Context) ([]*model.Post, error) {
	posts := make([]*model.Post, 0)

	for _, post := range mem.posts {
		newPost := post

		flatComments, err := mem.Comments(ctx, int(newPost.ID))
		if err != nil {
			log.Println(err)
		}

		commentMap := utils.BuildCommentTree(flatComments)

		var rootComments []*model.Comment

		for _, comment := range flatComments {
			if comment.ParentID == nil {
				rootComments = append(rootComments, commentMap[comment.ID])
			}
		}

		newPost.Comments = rootComments
		posts = append(posts, post)
	}

	return posts, nil
}

func (mem *InMemory) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	mem.commentIDs += 1
	mem.comments[mem.commentIDs] = &model.Comment{
		ID:        mem.commentIDs,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: time.Now().String(),
	}

	return mem.comments[mem.commentIDs], nil
}

func (mem *InMemory) Comments(ctx context.Context, postID int) ([]*model.Comment, error) {
	comments := make([]*model.Comment, 0)

	for _, comment := range mem.comments {
		if comment.PostID == int32(postID) {
			comments = append(comments, comment)
		}
	}

	sort.Slice(comments, func(i, j int) bool {
		if comments[i].PostID == comments[j].PostID {
			return comments[i].CreatedAt < comments[j].CreatedAt
		} else {
			return comments[i].PostID < comments[j].PostID
		}
	})

	return comments, nil
}
