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

func newSessionStore(db *sqlx.DB) *sessionStore {
	return &sessionStore{
		db: db,
	}
}

type sessionStore struct {
	db *sqlx.DB
}

type sqlDataSession struct {
	ID             int32     `db:"id"`
	Namespace      string    `db:"namespace"`
	DeviceID       string    `db:"device_id"`
	DeviceURI      string    `db:"device_uri"`
	SessionTimeout int       `db:"session_timeout"`
	LastMessageAt  time.Time `db:"last_message_at"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

var sqlParamsSession = []string{
	"id",
	"namespace",
	"device_id",
	"device_uri",
	"session_timeout",
	"last_message_at",
	"created_at",
	"updated_at",
}

func (d *sqlDataSession) Scan(m *model.Session) error {
	var createdAt, updatedAt = m.CreatedAt, m.UpdatedAt

	if m.CreatedAt.IsZero() {
		createdAt = time.Now().Round(time.Second).UTC()
	}

	if m.UpdatedAt.IsZero() {
		updatedAt = time.Now().Round(time.Second).UTC()
	}

	d.ID = m.ID
	d.Namespace = m.Namespace
	d.DeviceID = m.DeviceID
	d.DeviceURI = m.DeviceURI
	d.SessionTimeout = m.SessionTimeout
	d.LastMessageAt = m.LastMessageAt
	d.CreatedAt = createdAt
	d.UpdatedAt = updatedAt

	return nil
}

func (d *sqlDataSession) Model() (*model.Session, error) {
	m := &model.Session{
		ID:             d.ID,
		Namespace:      d.Namespace,
		DeviceID:       d.DeviceID,
		DeviceURI:      d.DeviceURI,
		SessionTimeout: d.SessionTimeout,
		LastMessageAt:  d.LastMessageAt,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}

	return m, nil
}

func (s *sessionStore) FetchAll() (map[int32]model.Session, error) {
	return fetchAllSessions(s.db)
}

func (s *sessionStore) FindByID(id int32) (*model.Session, error) {
	return findSessionByID(s.db, id)
}

func (s *sessionStore) FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Session, error) {
	return findSessionByNamespaceAndDeviceID(s.db, namespace, deviceID)
}

func (s *sessionStore) Create(m *model.Session) error {
	return createSession(s.db, m)
}

func (s *sessionStore) Update(m *model.Session) error {
	return updateSession(s.db, m)
}

func (s *sessionStore) Delete(id int32) error {
	return deleteSession(s.db, id)
}

func fetchAllSessions(db *sqlx.DB) (map[int32]model.Session, error) {
	rows := make([]sqlDataSession, 0)
	models := make(map[int32]model.Session)

	query := "SELECT * FROM sessions"
	if err := db.Select(&rows, query); err != nil {
		return nil, errors.Wrap(err, "failed to fetch all sessions")
	}

	for _, d := range rows {
		m, err := d.Model()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert SQL data to session model")
		}

		models[d.ID] = *m
	}

	return models, nil
}

func findSessionByID(db *sqlx.DB, id int32) (*model.Session, error) {
	d := sqlDataSession{}
	query := "SELECT * FROM sessions WHERE id=$1"
	if err := db.Get(&d, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to find session")
	}

	return d.Model()
}

func findSessionByNamespaceAndDeviceID(db *sqlx.DB, namespace, deviceID string) (*model.Session, error) {
	d := sqlDataSession{}
	query := "SELECT * FROM sessions WHERE namespace=$1 AND device_id=$2"
	if err := db.Get(&d, query, namespace, deviceID); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to find session")
	}

	return d.Model()
}

func createSession(db *sqlx.DB, m *model.Session) error {
	if m.SessionTimeout == 0 {
		m.SessionTimeout = 120
	}

	d := sqlDataSession{}
	if err := d.Scan(m); err != nil {
		return errors.Wrap(err, "failed to convert session model to SQL data")
	}

	// Remove the id column because it's of SQL type serial
	sqlParamsWithoutID := make([]string, 0)
	for _, s := range sqlParamsSession {
		if s != "id" {
			sqlParamsWithoutID = append(sqlParamsWithoutID, s)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO sessions (%s) VALUES (%s) RETURNING id",
		strings.Join(sqlParamsWithoutID, ", "),
		":"+strings.Join(sqlParamsWithoutID, ", :"),
	)
	rows, err := db.NamedQuery(query, d)
	if err != nil {
		return errors.Wrap(err, "failed to created session")
	}
	if rows.Next() {
		rows.Scan(&m.ID)
	}

	return nil
}

func updateSession(db *sqlx.DB, m *model.Session) error {
	if _, err := findSessionByID(db, m.ID); err != nil {
		return err
	}

	// Set the UpdateAt date to now
	m.UpdatedAt = time.Now().Round(time.Second).UTC()

	d := sqlDataSession{}
	if err := d.Scan(m); err != nil {
		return errors.Wrap(err, "failed to convert session model to SQL data")
	}

	var queryParams []string
	for _, param := range sqlParamsSession {
		queryParams = append(queryParams, fmt.Sprintf("%s=:%s", param, param))
	}
	query := fmt.Sprintf("UPDATE sessions SET %s WHERE id=:id", strings.Join(queryParams, ", "))
	if _, err := db.NamedExec(query, d); err != nil {
		return errors.Wrap(err, "failed to update session")
	}

	return nil
}

func deleteSession(db *sqlx.DB, id int32) error {
	query := "DELETE FROM sessions WHERE id=$1"
	_, err := db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	return nil
}
