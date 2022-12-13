package types

import "github.com/gorilla/websocket"

const (
	CREATE = iota
	LIKE   = iota
	DELETE = iota
)

type Connection struct {
	ID       string
	Username string
	Conn     websocket.Conn
}

type Chat struct {
	Name        string
	CreatedAt   string
	Duration    int
	Messages    []Message
	Connections []Connection
}

type Message struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Text     string   `json:"text"`
	File     string   `json:"file"`
	SentBy   string   `json:"sentBy"`
	SentAt   string   `json:"sentAt"`
	Likes    []string `json:"likes"`
}

type WS_Signal struct {
	EventType int     `json:"eventType"`
	Message   Message `json:"message"`
}

func (conn Connection) GetId() string {
	return conn.ID
}

func (message Message) GetId() string {
	return message.ID
}