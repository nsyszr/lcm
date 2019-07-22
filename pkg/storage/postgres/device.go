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

func newDeviceStore(db *sqlx.DB) *deviceStore {
	return &deviceStore{
		db: db,
	}
}

type deviceStore struct {
	db *sqlx.DB
}

type sqlDataDevice struct {
	ID             int32     `db:"id"`
	Namespace      string    `db:"namespace"`
	DeviceID       string    `db:"device_id"`
	DeviceURI      string    `db:"device_uri"`
	SessionTimeout int       `db:"session_timeout"`
	PingInterval   int       `db:"ping_interval"`
	PongTimeout    int       `db:"pong_timeout"`
	EventsTopic    string    `db:"events_topic"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

var sqlParamsDevice = []string{
	"id",
	"namespace",
	"device_id",
	"device_uri",
	"session_timeout",
	"ping_interval",
	"pong_timeout",
	"events_topic",
	"created_at",
	"updated_at",
}

func (d *sqlDataDevice) Scan(m *model.Device) error {
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
	d.PingInterval = m.PingInterval
	d.PongTimeout = m.PongTimeout
	d.EventsTopic = m.EventsTopic
	d.CreatedAt = createdAt
	d.UpdatedAt = updatedAt

	return nil
}

func (d *sqlDataDevice) Model() (*model.Device, error) {
	m := &model.Device{
		ID:             d.ID,
		Namespace:      d.Namespace,
		DeviceID:       d.DeviceID,
		DeviceURI:      d.DeviceURI,
		SessionTimeout: d.SessionTimeout,
		PingInterval:   d.PingInterval,
		PongTimeout:    d.PongTimeout,
		EventsTopic:    d.EventsTopic,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}

	return m, nil
}

func (s *deviceStore) FetchAll() (map[int32]model.Device, error) {
	return fetchAllDevices(s.db)
}

func (s *deviceStore) FindByID(id int32) (*model.Device, error) {
	return findDeviceByID(s.db, id)
}

func (s *deviceStore) Create(m *model.Device) error {
	return createDevice(s.db, m)
}

func (s *deviceStore) Delete(id int32) error {
	return deleteDevice(s.db, id)
}

func (s *deviceStore) FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Device, error) {
	return findDeviceByNamespaceAndDeviceID(s.db, namespace, deviceID)
}

func fetchAllDevices(db *sqlx.DB) (map[int32]model.Device, error) {
	rows := make([]sqlDataDevice, 0)
	models := make(map[int32]model.Device)

	query := "SELECT * FROM devices ORDER BY id"
	if err := db.Select(&rows, query); err != nil {
		return nil, errors.Wrap(err, "failed to fetch all devuces")
	}

	for _, d := range rows {
		m, err := d.Model()
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert SQL data to device model")
		}

		models[d.ID] = *m
	}

	return models, nil
}

func findDeviceByID(db *sqlx.DB, id int32) (*model.Device, error) {
	d := sqlDataDevice{}
	query := "SELECT * FROM devices WHERE id=$1"
	if err := db.Get(&d, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to find device")
	}

	return d.Model()
}

func findDeviceByNamespaceAndDeviceID(db *sqlx.DB, namespace, deviceID string) (*model.Device, error) {
	d := sqlDataDevice{}
	query := "SELECT * FROM devices WHERE namespace=$1 AND device_id=$2"
	if err := db.Get(&d, query, namespace, deviceID); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to find device")
	}

	return d.Model()
}

func createDevice(db *sqlx.DB, m *model.Device) error {
	// Set default values
	if m.SessionTimeout == 0 {
		m.SessionTimeout = 120
	}
	if m.PingInterval == 0 {
		m.PingInterval = 104
	}
	if m.PongTimeout == 0 {
		m.PongTimeout = 16
	}
	if m.EventsTopic == "" {
		m.EventsTopic = "deviceevent"
	}

	// TODO(DGL) Add validation, eg. PingInterval not greater than SessionTimeout

	d := sqlDataDevice{}
	if err := d.Scan(m); err != nil {
		return errors.Wrap(err, "failed to convert device model to SQL data")
	}

	// Remove the id column because it's of SQL type serial
	sqlParamsWithoutID := make([]string, 0)
	for _, s := range sqlParamsDevice {
		if s != "id" {
			sqlParamsWithoutID = append(sqlParamsWithoutID, s)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO devices (%s) VALUES (%s) RETURNING id",
		strings.Join(sqlParamsWithoutID, ", "),
		":"+strings.Join(sqlParamsWithoutID, ", :"),
	)
	rows, err := db.NamedQuery(query, d)
	if err != nil {
		return errors.Wrap(err, "failed to created device")
	}
	if rows.Next() {
		rows.Scan(&m.ID)
	}

	return nil
}

func deleteDevice(db *sqlx.DB, id int32) error {
	query := "DELETE FROM devices WHERE id=$1"
	_, err := db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	return nil
}
