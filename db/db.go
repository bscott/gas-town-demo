package db

import (
	"database/sql"
	"embed"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// Channel represents a chat channel
type Channel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// InitDB initializes the database and creates tables
func InitDB(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}

	if _, err := sqlDB.Exec(string(schema)); err != nil {
		return nil, err
	}

	return &DB{sqlDB}, nil
}

// CreateChannel creates a new channel
func (db *DB) CreateChannel(name string) (*Channel, error) {
	channel := &Channel{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
	}

	_, err := db.Exec(
		"INSERT INTO channels (id, name, created_at) VALUES (?, ?, ?)",
		channel.ID, channel.Name, channel.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannel retrieves a channel by ID
func (db *DB) GetChannel(id string) (*Channel, error) {
	channel := &Channel{}
	err := db.QueryRow(
		"SELECT id, name, created_at FROM channels WHERE id = ?",
		id,
	).Scan(&channel.ID, &channel.Name, &channel.CreatedAt)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// GetChannelByName retrieves a channel by name
func (db *DB) GetChannelByName(name string) (*Channel, error) {
	channel := &Channel{}
	err := db.QueryRow(
		"SELECT id, name, created_at FROM channels WHERE name = ?",
		name,
	).Scan(&channel.ID, &channel.Name, &channel.CreatedAt)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// ListChannels returns all channels
func (db *DB) ListChannels() ([]Channel, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM channels ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

// DeleteChannel deletes a channel by ID
func (db *DB) DeleteChannel(id string) error {
	_, err := db.Exec("DELETE FROM channels WHERE id = ?", id)
	return err
}

// CreateMessage creates a new message in a channel
func (db *DB) CreateMessage(channelID, author, content string) (*Message, error) {
	msg := &Message{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		Author:    author,
		Content:   content,
		CreatedAt: time.Now(),
	}

	_, err := db.Exec(
		"INSERT INTO messages (id, channel_id, author, content, created_at) VALUES (?, ?, ?, ?, ?)",
		msg.ID, msg.ChannelID, msg.Author, msg.Content, msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// GetMessage retrieves a message by ID
func (db *DB) GetMessage(id string) (*Message, error) {
	msg := &Message{}
	err := db.QueryRow(
		"SELECT id, channel_id, author, content, created_at FROM messages WHERE id = ?",
		id,
	).Scan(&msg.ID, &msg.ChannelID, &msg.Author, &msg.Content, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// ListMessages returns messages for a channel, ordered by creation time
func (db *DB) ListMessages(channelID string, limit int) ([]Message, error) {
	rows, err := db.Query(
		"SELECT id, channel_id, author, content, created_at FROM messages WHERE channel_id = ? ORDER BY created_at ASC LIMIT ?",
		channelID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.Author, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// DeleteMessage deletes a message by ID
func (db *DB) DeleteMessage(id string) error {
	_, err := db.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}
