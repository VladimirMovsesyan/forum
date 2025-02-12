package pubsub

import (
	"sync"

	"github.com/VladimirMovsesyan/forum/internal/domain/model"
)

type PubSub struct {
	mu          sync.Mutex
	subscribers map[int32][]chan *model.Comment
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[int32][]chan *model.Comment),
	}
}

func (ps *PubSub) Subscribe(postID int32) chan *model.Comment {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan *model.Comment, 1)
	ps.subscribers[postID] = append(ps.subscribers[postID], ch)

	return ch
}

func (ps *PubSub) Publish(postID int32, comment *model.Comment) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, ch := range ps.subscribers[postID] {
		ch <- comment
	}
}

func (ps *PubSub) Unsubscribe(postID int32, ch chan *model.Comment) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subscribers := ps.subscribers[postID]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			ps.subscribers[postID] = append(subscribers[:i], subscribers[i+1:]...)

			close(ch)

			break
		}
	}
}
