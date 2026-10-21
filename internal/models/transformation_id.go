package models

import (
	"database/sql"
	"fmt"
	"time"
)

type TransformationID struct {
	ID               int
	RequestID        int
	Label            string
	TransformationID int
	CreatedBy        int
	CreatorName      string
	CreatedAt        time.Time
}

type TransformationIDStore struct {
	db *sql.DB
	dbHelper
}

func NewTransformationIDStore(db *sql.DB, driver string) *TransformationIDStore {
	return &TransformationIDStore{db: db, dbHelper: newHelper(driver)}
}

func (s *TransformationIDStore) Add(requestID int, label string, transformationID int, userID int) error {
	var createdBy interface{}
	if userID != 0 {
		createdBy = userID
	}
	_, err := s.db.Exec(s.rebind(`
		INSERT INTO transformation_ids (request_id, label, transformation_id, created_by)
		VALUES (?, ?, ?, ?)`),
		requestID, label, transformationID, createdBy,
	)
	return err
}

func (s *TransformationIDStore) GetByRequestID(requestID int) ([]*TransformationID, error) {
	rows, err := s.db.Query(s.rebind(`
		SELECT p.id, p.request_id, p.label, p.transformation_id, COALESCE(p.created_by, 0),
		       COALESCE(u.display_name, ''), p.created_at
		FROM transformation_ids p
		LEFT JOIN users u ON u.id = p.created_by
		WHERE p.request_id = ?
		ORDER BY p.created_at ASC`), requestID)
	if err != nil {
		return nil, fmt.Errorf("query transformation ids: %w", err)
	}
	defer rows.Close()

	var ids []*TransformationID
	for rows.Next() {
		var p TransformationID
		if err := rows.Scan(&p.ID, &p.RequestID, &p.Label, &p.TransformationID, &p.CreatedBy, &p.CreatorName, timeVal{&p.CreatedAt}); err != nil {
			return nil, err
		}
		ids = append(ids, &p)
	}
	return ids, rows.Err()
}

func (s *TransformationIDStore) GetByID(id int) (*TransformationID, error) {
	var p TransformationID
	err := s.db.QueryRow(s.rebind(`
		SELECT id, request_id, label, transformation_id, COALESCE(created_by, 0), created_at
		FROM transformation_ids WHERE id = ?`), id).
		Scan(&p.ID, &p.RequestID, &p.Label, &p.TransformationID, &p.CreatedBy, timeVal{&p.CreatedAt})
	return &p, err
}

func (s *TransformationIDStore) Delete(id int) error {
	_, err := s.db.Exec(s.rebind("DELETE FROM transformation_ids WHERE id = ?"), id)
	return err
}
