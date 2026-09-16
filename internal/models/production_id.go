package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ProductionID struct {
	ID           int
	RequestID    int
	Label        string
	ProductionID int
	CreatedBy    int
	CreatorName  string
	CreatedAt    time.Time
}

type ProductionIDStore struct {
	db *sql.DB
	dbHelper
}

func NewProductionIDStore(db *sql.DB, driver string) *ProductionIDStore {
	return &ProductionIDStore{db: db, dbHelper: newHelper(driver)}
}

func (s *ProductionIDStore) Add(requestID int, label string, productionID int, userID int) error {
	var createdBy interface{}
	if userID != 0 {
		createdBy = userID
	}
	_, err := s.db.Exec(s.rebind(`
		INSERT INTO production_ids (request_id, label, production_id, created_by)
		VALUES (?, ?, ?, ?)`),
		requestID, label, productionID, createdBy,
	)
	return err
}

func (s *ProductionIDStore) GetByRequestID(requestID int) ([]*ProductionID, error) {
	rows, err := s.db.Query(s.rebind(`
		SELECT p.id, p.request_id, p.label, p.production_id, COALESCE(p.created_by, 0),
		       COALESCE(u.display_name, ''), p.created_at
		FROM production_ids p
		LEFT JOIN users u ON u.id = p.created_by
		WHERE p.request_id = ?
		ORDER BY p.created_at ASC`), requestID)
	if err != nil {
		return nil, fmt.Errorf("query production ids: %w", err)
	}
	defer rows.Close()

	var ids []*ProductionID
	for rows.Next() {
		var p ProductionID
		if err := rows.Scan(&p.ID, &p.RequestID, &p.Label, &p.ProductionID, &p.CreatedBy, &p.CreatorName, timeVal{&p.CreatedAt}); err != nil {
			return nil, err
		}
		ids = append(ids, &p)
	}
	return ids, rows.Err()
}

func (s *ProductionIDStore) GetByID(id int) (*ProductionID, error) {
	var p ProductionID
	err := s.db.QueryRow(s.rebind(`
		SELECT id, request_id, label, production_id, COALESCE(created_by, 0), created_at
		FROM production_ids WHERE id = ?`), id).
		Scan(&p.ID, &p.RequestID, &p.Label, &p.ProductionID, &p.CreatedBy, timeVal{&p.CreatedAt})
	return &p, err
}

func (s *ProductionIDStore) Delete(id int) error {
	_, err := s.db.Exec(s.rebind("DELETE FROM production_ids WHERE id = ?"), id)
	return err
}
