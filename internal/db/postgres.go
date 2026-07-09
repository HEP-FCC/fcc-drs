//go:build prod

package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Init() (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	sqldb, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db := &DB{DB: sqldb, driverName: "postgres"}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id           SERIAL PRIMARY KEY,
			username     TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			email        TEXT NOT NULL DEFAULT '',
			role         TEXT NOT NULL DEFAULT 'requester',
			created_at   TIMESTAMPTZ DEFAULT NOW(),
			last_login   TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS dataset_requests (
			id                 SERIAL PRIMARY KEY,
			title              TEXT NOT NULL,
			description        TEXT DEFAULT '',
			requester_name     TEXT NOT NULL,
			requester_username TEXT NOT NULL DEFAULT '',
			requester_email    TEXT DEFAULT '',
			dataset_type       TEXT DEFAULT 'simulation',
			use_case           TEXT DEFAULT 'physics_analysis',
			status             TEXT DEFAULT 'pending',
			priority           TEXT DEFAULT 'medium',
			estimated_size     TEXT DEFAULT '',
			statistics         TEXT DEFAULT '',
			target_campaign    TEXT DEFAULT '',
			key4hep_stack      TEXT DEFAULT '',
			format             TEXT DEFAULT '',
			due_date           TEXT DEFAULT '',
			notes              TEXT DEFAULT '',
			tags               TEXT DEFAULT '',
			created_by         INTEGER REFERENCES users(id),
			assigned_to        INTEGER REFERENCES users(id),
			physics_approval   TEXT NOT NULL DEFAULT '',
			resources_approval TEXT NOT NULL DEFAULT '',
			created_at         TIMESTAMPTZ DEFAULT NOW(),
			updated_at         TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS request_activity (
			id         SERIAL PRIMARY KEY,
			request_id INTEGER NOT NULL REFERENCES dataset_requests(id) ON DELETE CASCADE,
			user_id    INTEGER REFERENCES users(id),
			type       TEXT NOT NULL DEFAULT 'comment',
			body       TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE OR REPLACE FUNCTION update_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
		`CREATE TABLE IF NOT EXISTS request_relations (
			id         SERIAL PRIMARY KEY,
			from_id    INTEGER NOT NULL REFERENCES dataset_requests(id) ON DELETE CASCADE,
			to_id      INTEGER NOT NULL REFERENCES dataset_requests(id) ON DELETE CASCADE,
			type       TEXT NOT NULL DEFAULT 'related',
			created_by INTEGER REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(from_id, to_id, type)
		)`,
		`CREATE TABLE IF NOT EXISTS generator_cards (
			id          SERIAL PRIMARY KEY,
			request_id  INTEGER NOT NULL REFERENCES dataset_requests(id) ON DELETE CASCADE,
			filename    TEXT NOT NULL,
			size        BIGINT NOT NULL DEFAULT 0,
			content     BYTEA NOT NULL,
			uploaded_by INTEGER REFERENCES users(id),
			created_at  TIMESTAMPTZ DEFAULT NOW()
		)`,
		`ALTER TABLE dataset_requests ADD COLUMN IF NOT EXISTS statistics TEXT DEFAULT ''`,
		`ALTER TABLE dataset_requests ADD COLUMN IF NOT EXISTS target_campaign TEXT DEFAULT ''`,
		`ALTER TABLE dataset_requests ADD COLUMN IF NOT EXISTS key4hep_stack TEXT DEFAULT ''`,
		`DROP TRIGGER IF EXISTS update_timestamp ON dataset_requests`,
		`CREATE TRIGGER update_timestamp
			BEFORE UPDATE ON dataset_requests
			FOR EACH ROW EXECUTE FUNCTION update_updated_at()`,
		`UPDATE users SET role = 'coordinator' WHERE role = 'manager'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar BYTEA`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_mime TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS coordinator_groups (
			id          SERIAL PRIMARY KEY,
			name        TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS coordinator_group_members (
			group_id INTEGER NOT NULL REFERENCES coordinator_groups(id) ON DELETE CASCADE,
			user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY (group_id, user_id)
		)`,
		`ALTER TABLE dataset_requests ADD COLUMN IF NOT EXISTS assigned_group_id INTEGER REFERENCES coordinator_groups(id)`,
		`INSERT INTO coordinator_groups (name) VALUES
			('BSM physics'),
			('Electroweak physics'),
			('FCC-hh physics'),
			('Flavour physics'),
			('Global fits & EFT'),
			('Higgs physics'),
			('Precision calculations'),
			('QCD and photon-photon physics'),
			('Top-quark physics'),
			('Computing Resources'),
			('Core SW/Key4hep, Releases'),
			('Digitization and Reconstruction Software'),
			('Documentation and Trainings'),
			('Geometry, Simulation'),
			('Interaction Region, Beam Backgrounds'),
			('MC Productions, GRID Tools'),
			('Analysis Tools'),
			('High-level reconstruction'),
			('Monte Carlo tools')
		ON CONFLICT (name) DO NOTHING`,
		`UPDATE dataset_requests dr
			SET assigned_group_id = cg.id
			FROM coordinator_groups cg
			WHERE cg.name = dr.working_group
			  AND dr.working_group != ''
			  AND dr.assigned_group_id IS NULL`,
		`ALTER TABLE dataset_requests DROP COLUMN IF EXISTS working_group`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	return tx.Commit()
}
