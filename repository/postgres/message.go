package repository

import (
	"database/sql"

	"github.com/allisson/hammer"
	"github.com/jmoiron/sqlx"
)

// Message is a implementation of hammer.MessageRepository
type Message struct {
	db *sqlx.DB
}

// Find returns hammer.Message by id
func (m *Message) Find(id string) (hammer.Message, error) {
	message := hammer.Message{}
	findOptions := hammer.FindOptions{
		FindFilters: []hammer.FindFilter{
			{
				FieldName: "id",
				Operator:  "=",
				Value:     id,
			},
		},
	}
	sql, args := buildSQLQuery("messages", findOptions)
	err := m.db.Get(&message, sql, args...)
	return message, err
}

// FindAll returns []hammer.Message by limit and offset
func (m *Message) FindAll(findOptions hammer.FindOptions) ([]hammer.Message, error) {
	messages := []hammer.Message{}
	sql, args := buildSQLQuery("messages", findOptions)
	err := m.db.Select(&messages, sql, args...)
	return messages, err
}

// Store a hammer.Message on database (create or update)
func (m *Message) Store(tx hammer.TxRepository, message *hammer.Message) error {
	_, err := m.Find(message.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return tx.Exec(sqlMessageCreate, message)
		}
		return err
	}
	return tx.Exec(sqlMessageUpdate, message)
}

// Delete a hammer.Message on database
func (m *Message) Delete(tx hammer.TxRepository, id string) error {
	_, err := m.Find(id)
	if err != nil {
		return err
	}
	return tx.Exec(sqlMessageDelete, map[string]interface{}{"id": id})
}

// NewMessage returns a new Message with db connection
func NewMessage(db *sqlx.DB) Message {
	return Message{db: db}
}
