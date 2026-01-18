// Package conversation provides conversation management for multi-turn chat.
package conversation

import (
	"time"

	"agent-desktop/internal/llm"

	"github.com/google/uuid"
)

// Conversation represents a multi-turn conversation with the agent.
type Conversation struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []llm.Message `json:"messages"`
}

// Summary is a lightweight representation of a conversation for listing.
type Summary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TurnCount int       `json:"turn_count"`
}

// New creates a new conversation with a generated ID and default title.
func New() *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New().String(),
		Title:     "New Conversation",
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []llm.Message{},
	}
}

// AddMessage appends a message to the conversation and updates the timestamp.
func (c *Conversation) AddMessage(msg llm.Message) {
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// TurnCount returns the number of user messages (turns) in the conversation.
func (c *Conversation) TurnCount() int {
	count := 0
	for _, msg := range c.Messages {
		if msg.Role == "user" {
			count++
		}
	}
	return count
}

// ToSummary creates a Summary from this conversation.
func (c *Conversation) ToSummary() Summary {
	return Summary{
		ID:        c.ID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		TurnCount: c.TurnCount(),
	}
}
