package domain

import "time"

// PostStatsEventType represents the stats event type.
type PostStatsEventType string

const (
	PostStatsEventLike      PostStatsEventType = "like"
	PostStatsEventUnlike    PostStatsEventType = "unlike"
	PostStatsEventCollect   PostStatsEventType = "collect"
	PostStatsEventUncollect PostStatsEventType = "uncollect"
	PostStatsEventRead      PostStatsEventType = "read"
)

// PostStatsEvent is published to MQ for async counter updates.
type PostStatsEvent struct {
	EventId string             `json:"event_id"`
	Type    PostStatsEventType `json:"type"`
	PostId  int64              `json:"post_id"`
	UserId  int64              `json:"user_id"`
	Ts      int64              `json:"ts"`
}

func NewPostStatsEvent(eventId string, eventType PostStatsEventType, postId, userId int64) PostStatsEvent {
	return PostStatsEvent{
		EventId: eventId,
		Type:    eventType,
		PostId:  postId,
		UserId:  userId,
		Ts:      time.Now().Unix(),
	}
}
