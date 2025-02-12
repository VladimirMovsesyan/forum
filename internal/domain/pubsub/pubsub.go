package pubsub

import (
	"github.com/VladimirMovsesyan/forum/internal/domain/model"
	"sync"
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

func (ps *PubSub) Subscribe(postId int32) chan *model.Comment {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan *model.Comment, 1)
	ps.subscribers[postId] = append(ps.subscribers[postId], ch)
	return ch
}

func (ps *PubSub) Publish(postId int32, comment *model.Comment) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, ch := range ps.subscribers[postId] {
		ch <- comment
	}
}

func (ps *PubSub) Unsubscribe(postId int32, ch chan *model.Comment) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subscribers := ps.subscribers[postId]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			ps.subscribers[postId] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}
