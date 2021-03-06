package repository

import (
	"database/sql"

	"github.com/allisson/hammer"
	"github.com/jmoiron/sqlx"
)

// Subscription is a implementation of hammer.SubscriptionRepository
type Subscription struct {
	db *sqlx.DB
}

// Find returns hammer.Subscription by id
func (s *Subscription) Find(id string) (hammer.Subscription, error) {
	subscription := hammer.Subscription{}
	findOptions := hammer.FindOptions{
		FindFilters: []hammer.FindFilter{
			{
				FieldName: "id",
				Operator:  "=",
				Value:     id,
			},
		},
	}
	sql, args := buildSQLQuery("subscriptions", findOptions)
	err := s.db.Get(&subscription, sql, args...)
	return subscription, err
}

// FindAll returns []hammer.Subscription by limit and offset
func (s *Subscription) FindAll(findOptions hammer.FindOptions) ([]hammer.Subscription, error) {
	subscriptions := []hammer.Subscription{}
	sql, args := buildSQLQuery("subscriptions", findOptions)
	err := s.db.Select(&subscriptions, sql, args...)
	return subscriptions, err
}

// Store a hammer.Subscription on database (create or update)
func (s *Subscription) Store(tx hammer.TxRepository, subscription *hammer.Subscription) error {
	_, err := s.Find(subscription.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return tx.Exec(sqlSubscriptionCreate, subscription)
		}
		return err
	}
	return tx.Exec(sqlSubscriptionUpdate, subscription)
}

// Delete a hammer.Subscription on database
func (s *Subscription) Delete(tx hammer.TxRepository, id string) error {
	_, err := s.Find(id)
	if err != nil {
		return err
	}
	return tx.Exec(sqlSubscriptionDelete, map[string]interface{}{"id": id})
}

// NewSubscription returns a new Subscription with db connection
func NewSubscription(db *sqlx.DB) Subscription {
	return Subscription{db: db}
}
