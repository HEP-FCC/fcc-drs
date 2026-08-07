package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Campaign struct {
	ID        int
	Name      string
	Tag       string
	Status    string // "open" | "closed"
	CreatedAt time.Time
	ClosedAt  time.Time // zero value if the campaign has never been closed
}

const campaignSelectCols = `id, name, tag, status, created_at, closed_at`

type CampaignStore struct {
	db *sql.DB
	dbHelper
}

func NewCampaignStore(db *sql.DB, driver string) *CampaignStore {
	return &CampaignStore{db: db, dbHelper: newHelper(driver)}
}

func scanCampaign(row scannable) (*Campaign, error) {
	c := &Campaign{}
	err := row.Scan(&c.ID, &c.Name, &c.Tag, &c.Status, timeVal{&c.CreatedAt}, timeVal{&c.ClosedAt})
	return c, err
}

func (s *CampaignStore) GetAll() ([]*Campaign, error) {
	rows, err := s.db.Query(`SELECT ` + campaignSelectCols + ` FROM campaigns ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (s *CampaignStore) GetOpen() ([]*Campaign, error) {
	rows, err := s.db.Query(`SELECT ` + campaignSelectCols + ` FROM campaigns WHERE status = 'open' ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query open campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

// GetAssignable returns open campaigns plus, if currentID refers to a closed
// campaign, that campaign too — so a request already assigned to a since-closed
// campaign still shows it as the selected option.
func (s *CampaignStore) GetAssignable(currentID int) ([]*Campaign, error) {
	campaigns, err := s.GetOpen()
	if err != nil {
		return nil, err
	}
	if currentID == 0 {
		return campaigns, nil
	}
	for _, c := range campaigns {
		if c.ID == currentID {
			return campaigns, nil
		}
	}
	current, err := s.GetByID(currentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return campaigns, nil
		}
		return nil, err
	}
	return append(campaigns, current), nil
}

func (s *CampaignStore) GetByID(id int) (*Campaign, error) {
	row := s.db.QueryRow(s.rebind(`SELECT `+campaignSelectCols+` FROM campaigns WHERE id = ?`), id)
	return scanCampaign(row)
}

func (s *CampaignStore) Create(name, tag string) error {
	_, err := s.db.Exec(s.rebind(`INSERT INTO campaigns (name, tag) VALUES (?, ?)`), name, tag)
	return err
}

// UpdateStatus sets the campaign's status, stamping closed_at when closing
// and clearing it when reopened.
func (s *CampaignStore) UpdateStatus(id int, status string) error {
	if status == "closed" {
		_, err := s.db.Exec(s.rebind(`UPDATE campaigns SET status=?, closed_at=CURRENT_TIMESTAMP WHERE id=?`), status, id)
		return err
	}
	_, err := s.db.Exec(s.rebind(`UPDATE campaigns SET status=?, closed_at=NULL WHERE id=?`), status, id)
	return err
}

func (s *CampaignStore) Delete(id int) error {
	_, err := s.db.Exec(s.rebind(`DELETE FROM campaigns WHERE id = ?`), id)
	return err
}
