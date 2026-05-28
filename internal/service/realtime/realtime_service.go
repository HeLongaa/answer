package realtime

import (
	"sync"
	"time"
)

const (
	EventQuestionCreated      = "question.created"
	EventQuestionFeatured     = "question.featured"
	EventPointsChanged        = "points.changed"
	EventTasksChanged         = "tasks.changed"
	EventFeaturedPostsChanged = "featured_posts.changed"
	EventAdminUsersChanged    = "admin.users.changed"
	EventTagChanged           = "tag.changed"
)

type Event struct {
	Type   string         `json:"type"`
	UserID string         `json:"user_id,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
	At     int64          `json:"at"`
}

type subscriber struct {
	userID string
	ch     chan Event
}

type Service struct {
	mu          sync.RWMutex
	subscribers map[chan Event]subscriber
}

func NewService() *Service {
	return &Service{subscribers: make(map[chan Event]subscriber)}
}

func (s *Service) Subscribe(userID string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	s.mu.Lock()
	s.subscribers[ch] = subscriber{userID: userID, ch: ch}
	s.mu.Unlock()

	return ch, func() {
		s.mu.Lock()
		if _, ok := s.subscribers[ch]; ok {
			delete(s.subscribers, ch)
			close(ch)
		}
		s.mu.Unlock()
	}
}

func (s *Service) Broadcast(eventType string, data map[string]any) {
	s.publish(Event{Type: eventType, Data: data, At: time.Now().Unix()})
}

func (s *Service) SendToUser(userID, eventType string, data map[string]any) {
	s.publish(Event{Type: eventType, UserID: userID, Data: data, At: time.Now().Unix()})
}

func (s *Service) publish(event Event) {
	if event.At == 0 {
		event.At = time.Now().Unix()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sub := range s.subscribers {
		if event.UserID != "" && event.UserID != sub.userID {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
}
