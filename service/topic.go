package service

import (
	"database/sql"
	"time"

	"github.com/allisson/hammer"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	// Set logger
	logger, _ = zap.NewProduction()
}

// Topic is a implementation of hammer.TopicService
type Topic struct {
	topicRepo     hammer.TopicRepository
	txFactoryRepo hammer.TxFactoryRepository
}

// Find returns hammer.Topic by id
func (t *Topic) Find(id string) (hammer.Topic, error) {
	return t.topicRepo.Find(id)
}

// FindAll returns []hammer.Topic by limit and offset
func (t *Topic) FindAll(findOptions hammer.FindOptions) ([]hammer.Topic, error) {
	return t.topicRepo.FindAll(findOptions)
}

// Create a hammer.Topic on repository
func (t *Topic) Create(topic *hammer.Topic) error {
	// Verify if object already exists
	_, err := t.topicRepo.Find(topic.ID)
	if err == nil {
		return hammer.ErrTopicAlreadyExists
	}

	// Create new topic
	tx, err := t.txFactoryRepo.New()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	topic.CreatedAt = now
	topic.UpdatedAt = now
	err = t.topicRepo.Store(tx, topic)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		rollback(tx, "topic-create-rollback")
		return err
	}

	return nil
}

// Update a hammer.Topic on repository
func (t *Topic) Update(topic *hammer.Topic) error {
	// Verify if object already exists
	topicFromRepo, err := t.topicRepo.Find(topic.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return hammer.ErrTopicDoesNotExists
		}
		return err
	}

	// Update topic
	tx, err := t.txFactoryRepo.New()
	if err != nil {
		return err
	}
	topic.ID = topicFromRepo.ID
	topic.CreatedAt = topicFromRepo.CreatedAt
	topic.UpdatedAt = time.Now().UTC()
	err = t.topicRepo.Store(tx, topic)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		rollback(tx, "topic-update-rollback")
		return err
	}

	return nil
}

// Delete a hammer.Topic on repository
func (t *Topic) Delete(id string) error {
	tx, err := t.txFactoryRepo.New()
	if err != nil {
		return err
	}

	err = t.topicRepo.Delete(tx, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		rollback(tx, "topic-delete-rollback")
		return err
	}

	return nil
}

// NewTopic returns a new Topic with topicRepo
func NewTopic(topicRepo hammer.TopicRepository, txFactoryRepo hammer.TxFactoryRepository) Topic {
	return Topic{
		topicRepo:     topicRepo,
		txFactoryRepo: txFactoryRepo,
	}
}
