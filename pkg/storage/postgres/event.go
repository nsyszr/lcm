package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nsyszr/lcm/pkg/model"
	"github.com/nsyszr/lcm/pkg/storage"
	"github.com/pkg/errors"
)

func newEventStore(db *sqlx.DB) *eventStore {
	return &eventStore{
		db: db,
	}
}

type eventStore struct {
	db *sqlx.DB
}

type sqlDataEvent struct {
	ID         int32     `db:"id"`
	Namespace  string    `db:"namespace"`
	SourceType string    `db:"source_type"`
	SourceID   string    `db:"source_id"`
	Topic      string    `db:"topic"`
	Timestamp  time.Time `db:"timestamp"`
	Details    string    `db:"details"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

var sqlParamsEvent = []string{
	"id",
	"namespace",
	"source_type",
	"source_id",
	"topic",
	"timestamp",
	"details",
	"created_at",
	"updated_at",
}

func (d *sqlDataEvent) Scan(m *model.Event) error {
	var createdAt, updatedAt = m.CreatedAt, m.UpdatedAt

	if m.CreatedAt.IsZero() {
		createdAt = time.Now().Round(time.Second).UTC()
	}

	if m.UpdatedAt.IsZero() {
		updatedAt = time.Now().Round(time.Second).UTC()
	}

	d.ID = m.ID
	d.Namespace = m.Namespace
	d.SourceType = m.SourceType
	d.SourceID = m.SourceID
	d.Topic = m.Topic
	d.Timestamp = m.Timestamp
	d.Details = m.Details
	d.CreatedAt = createdAt
	d.UpdatedAt = updatedAt

	return nil
}

func (d *sqlDataEvent) Model() (*model.Event, error) {
	m := &model.Event{
		ID:         d.ID,
		Namespace:  d.Namespace,
		SourceType: d.SourceType,
		SourceID:   d.SourceID,
		Topic:      d.Topic,
		Timestamp:  d.Timestamp,
		Details:    d.Details,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}

	return m, nil
}

func (s *eventStore) FetchAll() (map[int32]model.Event, error) {
	return fetchAllEvents(s.db)
}

func (s *eventStore) FindByID(id int32) (*model.Event, error) {
	return findEventByID(s.db, id)
}

func (s *eventStore) Create(m *model.Event) error {
	return createEvent(s.db, m)
}

func fetchAllEvents(db *sqlx.DB) (map[int32]model.Event, error) {
	rows := make([]sqlDataEvent, 0)
	models := make(map[int32]model.Event)

	query := "SELECT * FROM events"
	if err := db.Select(&rows, query); err != nil {
		return nil, errors.Wrap(err, "failed to fetch all events")
	}

	for _, d := range rows {
		m, err := d.Model()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert SQL data to event model")
		}

		models[d.ID] = *m
	}

	return models, nil
}

func findEventByID(db *sqlx.DB, id int32) (*model.Event, error) {
	d := sqlDataEvent{}
	query := "SELECT * FROM events WHERE id=$1"
	if err := db.Get(&d, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to find event")
	}

	return d.Model()
}

func createEvent(db *sqlx.DB, m *model.Event) error {
	d := sqlDataEvent{}
	if err := d.Scan(m); err != nil {
		return errors.Wrap(err, "failed to convert event model to SQL data")
	}

	// Remove the id column because it's of SQL type serial
	sqlParamsWithoutID := make([]string, 0)
	for _, s := range sqlParamsEvent {
		if s != "id" {
			sqlParamsWithoutID = append(sqlParamsWithoutID, s)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO events (%s) VALUES (%s)",
		strings.Join(sqlParamsWithoutID, ", "),
		":"+strings.Join(sqlParamsWithoutID, ", :"),
	)
	rows, err := db.NamedQuery(query, d)
	if err != nil {
		return errors.Wrap(err, "failed to created event")
	}
	if rows.Next() {
		rows.Scan(&m.ID)
	}

	return nil
}
