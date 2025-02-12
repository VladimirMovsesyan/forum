package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"github.com/VladimirMovsesyan/forum/internal/domain/utils"
)

func getReferer(x int32) *int32 {
	return &x
}

func TestBuildCommentTree(t *testing.T) {
	tests := []struct {
		name         string
		flatComments []*model.Comment
		want         map[int32]*model.Comment
	}{
		{
			name:         "empty",
			flatComments: []*model.Comment{},
			want:         map[int32]*model.Comment{},
		},
		{
			name: "single",
			flatComments: []*model.Comment{
				{
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
			want: map[int32]*model.Comment{
				1: {
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
		},
		{
			name: "many",
			flatComments: []*model.Comment{
				{
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				{
					ID:        2,
					PostID:    1,
					ParentID:  nil,
					Content:   "line2",
					Author:    "user2",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				{
					ID:        3,
					PostID:    1,
					ParentID:  nil,
					Content:   "line3",
					Author:    "user3",
					CreatedAt: time.Date(2029, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
			want: map[int32]*model.Comment{
				1: {
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				2: {
					ID:        2,
					PostID:    1,
					ParentID:  nil,
					Content:   "line2",
					Author:    "user2",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				3: {
					ID:        3,
					PostID:    1,
					ParentID:  nil,
					Content:   "line3",
					Author:    "user3",
					CreatedAt: time.Date(2029, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
		},
		{
			name: "deeper hierarchy",
			flatComments: []*model.Comment{
				{
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  []*model.Comment{},
				},
				{
					ID:        2,
					PostID:    1,
					ParentID:  getReferer(1),
					Content:   "line2",
					Author:    "user2",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				{
					ID:        3,
					PostID:    1,
					ParentID:  getReferer(1),
					Content:   "line3",
					Author:    "user3",
					CreatedAt: time.Date(2029, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
			want: map[int32]*model.Comment{
				1: {
					ID:        1,
					PostID:    1,
					ParentID:  nil,
					Content:   "line",
					Author:    "user",
					CreatedAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children: []*model.Comment{
						{
							ID:        2,
							PostID:    1,
							ParentID:  getReferer(1),
							Content:   "line2",
							Author:    "user2",
							CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local).String(),
							Children:  make([]*model.Comment, 0),
						},
						{
							ID:        3,
							PostID:    1,
							ParentID:  getReferer(1),
							Content:   "line3",
							Author:    "user3",
							CreatedAt: time.Date(2029, time.January, 1, 0, 0, 0, 0, time.Local).String(),
							Children:  make([]*model.Comment, 0),
						},
					},
				},
				2: {
					ID:        2,
					PostID:    1,
					ParentID:  getReferer(1),
					Content:   "line2",
					Author:    "user2",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
				3: {
					ID:        3,
					PostID:    1,
					ParentID:  getReferer(1),
					Content:   "line3",
					Author:    "user3",
					CreatedAt: time.Date(2029, time.January, 1, 0, 0, 0, 0, time.Local).String(),
					Children:  make([]*model.Comment, 0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.BuildCommentTree(tt.flatComments)
			require.Equal(t, len(tt.want), len(got))

			for k, v := range got {
				require.Equal(t, *tt.want[k], *v)
			}
		})
	}
}
