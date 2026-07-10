package models

import (
	"database/sql"
	"fmt"
)

// CoordinatorGroups is the canonical ordered list matching the coordinator_groups table.
var CoordinatorGroups = []struct {
	Category string
	Name     string
}{
	{"Physics Groups", "BSM physics"},
	{"Physics Groups", "Electroweak physics"},
	{"Physics Groups", "FCC-hh physics"},
	{"Physics Groups", "Flavour physics"},
	{"Physics Groups", "Global fits & EFT"},
	{"Physics Groups", "Higgs physics"},
	{"Physics Groups", "Precision calculations"},
	{"Physics Groups", "QCD and photon-photon physics"},
	{"Physics Groups", "Top-quark physics"},
	{"Software Groups", "Computing Resources"},
	{"Software Groups", "Core SW/Key4hep, Releases"},
	{"Software Groups", "Digitization and Reconstruction Software"},
	{"Software Groups", "Documentation and Trainings"},
	{"Software Groups", "Geometry, Simulation"},
	{"Software Groups", "Interaction Region, Beam Backgrounds"},
	{"Software Groups", "MC Productions, GRID Tools"},
	{"Shared Groups", "Analysis Tools"},
	{"Shared Groups", "High-level reconstruction"},
	{"Shared Groups", "Monte Carlo tools"},
}

type CoordinatorGroup struct {
	ID          int
	Name        string
	Description string
	Members     []*User
}

type CoordinatorGroupStore struct {
	db *sql.DB
	dbHelper
}

func NewCoordinatorGroupStore(db *sql.DB, driver string) *CoordinatorGroupStore {
	return &CoordinatorGroupStore{db: db, dbHelper: newHelper(driver)}
}

func (s *CoordinatorGroupStore) GetAll() ([]*CoordinatorGroup, error) {
	rows, err := s.db.Query(`SELECT id, name, description FROM coordinator_groups ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query coordinator groups: %w", err)
	}
	defer rows.Close()

	var groups []*CoordinatorGroup
	for rows.Next() {
		g := &CoordinatorGroup{}
		if err := rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, g := range groups {
		g.Members, err = s.getMembers(g.ID)
		if err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func (s *CoordinatorGroupStore) GetByName(name string) (*CoordinatorGroup, error) {
	row := s.db.QueryRow(s.rebind(`SELECT id, name, description FROM coordinator_groups WHERE name = ?`), name)
	g := &CoordinatorGroup{}
	if err := row.Scan(&g.ID, &g.Name, &g.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var err error
	g.Members, err = s.getMembers(g.ID)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *CoordinatorGroupStore) GetByID(id int) (*CoordinatorGroup, error) {
	row := s.db.QueryRow(s.rebind(`SELECT id, name, description FROM coordinator_groups WHERE id = ?`), id)
	g := &CoordinatorGroup{}
	if err := row.Scan(&g.ID, &g.Name, &g.Description); err != nil {
		return nil, err
	}
	var err error
	g.Members, err = s.getMembers(g.ID)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *CoordinatorGroupStore) getMembers(groupID int) ([]*User, error) {
	rows, err := s.db.Query(s.rebind(`
		SELECT u.id, u.username, u.display_name, u.preferred_name, u.email, u.role,
		       u.avatar IS NOT NULL AS has_avatar, u.created_at, u.last_login,
			       u.notify_new_requests, u.notify_status_changes, u.notify_comments
		FROM coordinator_group_members cgm
		JOIN users u ON u.id = cgm.user_id
		WHERE cgm.group_id = ?
		ORDER BY u.display_name`), groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, rows.Err()
}

func (s *CoordinatorGroupStore) Create(name, description string) error {
	_, err := s.db.Exec(s.rebind(`
		INSERT INTO coordinator_groups (name, description) VALUES (?, ?)`), name, description)
	return err
}

func (s *CoordinatorGroupStore) Delete(id int) error {
	_, err := s.db.Exec(s.rebind(`DELETE FROM coordinator_groups WHERE id = ?`), id)
	return err
}

func (s *CoordinatorGroupStore) AddMember(groupID, userID int) error {
	_, err := s.db.Exec(s.rebind(`
		INSERT INTO coordinator_group_members (group_id, user_id)
		VALUES (?, ?)
		ON CONFLICT (group_id, user_id) DO NOTHING`), groupID, userID)
	return err
}

func (s *CoordinatorGroupStore) RemoveMember(groupID, userID int) error {
	_, err := s.db.Exec(s.rebind(`
		DELETE FROM coordinator_group_members WHERE group_id = ? AND user_id = ?`), groupID, userID)
	return err
}

// GetMembershipsByUser returns a map of userID → groups that user belongs to.
// Groups in the map contain only ID/Name/Description (no recursive Members list).
func (s *CoordinatorGroupStore) GetMembershipsByUser() (map[int][]*CoordinatorGroup, error) {
	rows, err := s.db.Query(`
		SELECT cgm.user_id, g.id, g.name, g.description
		FROM coordinator_group_members cgm
		JOIN coordinator_groups g ON g.id = cgm.group_id
		ORDER BY cgm.user_id, g.id`)
	if err != nil {
		return nil, fmt.Errorf("query group memberships: %w", err)
	}
	defer rows.Close()

	result := map[int][]*CoordinatorGroup{}
	for rows.Next() {
		var userID int
		g := &CoordinatorGroup{}
		if err := rows.Scan(&userID, &g.ID, &g.Name, &g.Description); err != nil {
			return nil, err
		}
		result[userID] = append(result[userID], g)
	}
	return result, rows.Err()
}
