package utils

import "github.com/VladimirMovsesyan/forum/internal/domain/model"

func BuildCommentTree(flatComments []*model.Comment) map[int32]*model.Comment {
	commentMap := make(map[int32]*model.Comment)

	for i := range flatComments {
		comment := flatComments[i]
		comment.Children = []*model.Comment{}
		commentMap[comment.ID] = comment
	}

	for _, comment := range flatComments {
		if comment.ParentID != nil {
			parent := commentMap[*comment.ParentID]
			if parent != nil {
				parent.Children = append(parent.Children, commentMap[comment.ID])
			}
		}
	}

	return commentMap
}
