package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Role string

const (
	RoleRequester   Role = "requester"
	RoleCoordinator Role = "coordinator"
	RoleAdmin       Role = "admin"
)

type User struct {
	ID            int
	Username      string // preferred_username from OIDC
	DisplayName   string
	PreferredName string
	Email         string
	Role          Role
	HasAvatar     bool
	CreatedAt     time.Time
	LastLogin     time.Time
}

func (u *User) IsAdmin() bool {
	if u == nil {
		return false
	}
	return u.Role == RoleAdmin
}

// IsCoordinator returns true for both coordinators and admins (admins have all coordinator privileges).
func (u *User) IsCoordinator() bool {
	if u == nil {
		return false
	}
	return u.Role == RoleCoordinator || u.Role == RoleAdmin
}

func (u *User) IsRequester() bool {
	if u == nil {
		return false
	}
	return u.Role == RoleRequester
}

// DisplayedName returns PreferredName if set, otherwise DisplayName.
func (u *User) DisplayedName() string {
	if u == nil {
		return ""
	}
	if u.PreferredName != "" {
		return u.PreferredName
	}
	return u.DisplayName
}

// Initial returns the first character of the displayed name, safe for nil.
func (u *User) Initial() string {
	n := u.DisplayedName()
	if n == "" {
		return "?"
	}
	return string([]rune(n)[0])
}


type UserStore struct {
	db *sql.DB
	dbHelper
}

func NewUserStore(db *sql.DB, driver string) *UserStore {
	return &UserStore{db: db, dbHelper: newHelper(driver)}
}

// Upsert creates or updates a user on every SSO login, returning the current record.
func (r *UserStore) Upsert(username, displayName, email string, role Role) (*User, error) {
	_, err := r.db.Exec(r.rebind(`
		INSERT INTO users (username, display_name, email, role, last_login)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(username) DO UPDATE SET
			display_name = excluded.display_name,
			email        = excluded.email,
			last_login   = CURRENT_TIMESTAMP
	`), username, displayName, email, role)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	return r.GetByUsername(username)
}

func (r *UserStore) GetByUsername(username string) (*User, error) {
	row := r.db.QueryRow(r.rebind(`
		SELECT id, username, display_name, preferred_name, email, role,
		       avatar IS NOT NULL as has_avatar, created_at, last_login
		FROM users WHERE username = ?`), username)
	return scanUser(row)
}

func (r *UserStore) GetByID(id int) (*User, error) {
	row := r.db.QueryRow(r.rebind(`
		SELECT id, username, display_name, preferred_name, email, role,
		       avatar IS NOT NULL as has_avatar, created_at, last_login
		FROM users WHERE id = ?`), id)
	return scanUser(row)
}

// Sessions

func (r *UserStore) CreateSession(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		r.rebind(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`),
		token, userID, expiresAt,
	)
	return err
}

func (r *UserStore) GetSession(token string) (*User, error) {
	row := r.db.QueryRow(r.rebind(`
		SELECT u.id, u.username, u.display_name, u.preferred_name, u.email, u.role,
		       u.avatar IS NOT NULL as has_avatar, u.created_at, u.last_login
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ? AND s.expires_at > CURRENT_TIMESTAMP
	`), token)
	return scanUser(row)
}

func (r *UserStore) DeleteSession(token string) error {
	_, err := r.db.Exec(r.rebind(`DELETE FROM sessions WHERE id = ?`), token)
	return err
}

func (r *UserStore) GetCoordinators() ([]*User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, display_name, preferred_name, email, role,
		       avatar IS NOT NULL as has_avatar, created_at, last_login
		FROM users WHERE role IN ('coordinator', 'admin') ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var coordinators []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		coordinators = append(coordinators, u)
	}
	return coordinators, rows.Err()
}

func (r *UserStore) GetAll() ([]*User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, display_name, preferred_name, email, role,
		       avatar IS NOT NULL as has_avatar, created_at, last_login
		FROM users ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserStore) UpdateRole(userID int, role Role) error {
	_, err := r.db.Exec(r.rebind(`UPDATE users SET role = ? WHERE id = ?`), role, userID)
	return err
}

func (r *UserStore) UpdatePreferredName(userID int, name string) error {
	_, err := r.db.Exec(r.rebind(`UPDATE users SET preferred_name = ? WHERE id = ?`), name, userID)
	return err
}

func (r *UserStore) UpdateAvatar(userID int, data []byte, mime string) error {
	_, err := r.db.Exec(r.rebind(`UPDATE users SET avatar = ?, avatar_mime = ? WHERE id = ?`), data, mime, userID)
	return err
}

func (r *UserStore) DeleteAvatar(userID int) error {
	_, err := r.db.Exec(r.rebind(`UPDATE users SET avatar = NULL, avatar_mime = '' WHERE id = ?`), userID)
	return err
}

func (r *UserStore) GetAvatar(username string) ([]byte, string, error) {
	row := r.db.QueryRow(r.rebind(`SELECT avatar, avatar_mime FROM users WHERE username = ?`), username)
	var data []byte
	var mime string
	err := row.Scan(&data, &mime)
	return data, mime, err
}

func (r *UserStore) PurgeExpiredSessions() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`)
	return err
}

func scanUser(row scannable) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Username, &u.DisplayName, &u.PreferredName, &u.Email, &u.Role,
		&u.HasAvatar, timeVal{&u.CreatedAt}, timeVal{&u.LastLogin},
	)
	if u.PreferredName != "" {
		u.DisplayName = u.PreferredName
	}
	return &u, err
}
