package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", "./clariphish.db")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	return runMigrations()
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		api_key TEXT UNIQUE,
		role TEXT DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS campaigns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		status TEXT DEFAULT 'draft',
		created_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		launch_date TIMESTAMP,
		completed_date TIMESTAMP,
		template_id INTEGER,
		page_id INTEGER,
		smtp_id INTEGER,
		url TEXT,
		user_id INTEGER,
		FOREIGN KEY (template_id) REFERENCES templates(id),
		FOREIGN KEY (page_id) REFERENCES pages(id),
		FOREIGN KEY (smtp_id) REFERENCES smtp(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		subject TEXT NOT NULL,
		text TEXT,
		html TEXT,
		user_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		html TEXT NOT NULL,
		capture_credentials INTEGER DEFAULT 1,
		capture_passwords INTEGER DEFAULT 1,
		redirect_url TEXT,
		user_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		user_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT,
		last_name TEXT,
		email TEXT NOT NULL,
		position TEXT,
		group_id INTEGER,
		FOREIGN KEY (group_id) REFERENCES groups(id)
	);

	CREATE TABLE IF NOT EXISTS campaign_targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		campaign_id INTEGER,
		target_id INTEGER,
		rid TEXT UNIQUE,
		status TEXT DEFAULT 'scheduled',
		send_date TIMESTAMP,
		FOREIGN KEY (campaign_id) REFERENCES campaigns(id),
		FOREIGN KEY (target_id) REFERENCES targets(id)
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		campaign_id INTEGER,
		campaign_target_id INTEGER,
		email TEXT,
		time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		message TEXT,
		details TEXT,
		FOREIGN KEY (campaign_id) REFERENCES campaigns(id),
		FOREIGN KEY (campaign_target_id) REFERENCES campaign_targets(id)
	);

	CREATE TABLE IF NOT EXISTS smtp (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		username TEXT,
		password TEXT,
		from_address TEXT NOT NULL,
		user_id INTEGER,
		ignore_cert_errors INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS assessments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		deadline TIMESTAMP,
		is_published INTEGER DEFAULT 0,
		created_by INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS questions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assessment_id INTEGER NOT NULL,
		question_text TEXT NOT NULL,
		question_order INTEGER DEFAULT 0,
		points INTEGER DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (assessment_id) REFERENCES assessments(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS answer_options (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		question_id INTEGER NOT NULL,
		option_text TEXT NOT NULL,
		is_correct INTEGER DEFAULT 0,
		option_order INTEGER DEFAULT 0,
		FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS user_assessment_attempts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		assessment_id INTEGER NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		score INTEGER DEFAULT 0,
		total_points INTEGER DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (assessment_id) REFERENCES assessments(id)
	);

	CREATE TABLE IF NOT EXISTS user_responses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		attempt_id INTEGER NOT NULL,
		question_id INTEGER NOT NULL,
		selected_option_id INTEGER NOT NULL,
		is_correct INTEGER DEFAULT 0,
		points_earned INTEGER DEFAULT 0,
		answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (attempt_id) REFERENCES user_assessment_attempts(id) ON DELETE CASCADE,
		FOREIGN KEY (question_id) REFERENCES questions(id),
		FOREIGN KEY (selected_option_id) REFERENCES answer_options(id)
	);

	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_events_campaign ON events(campaign_id);
	CREATE INDEX IF NOT EXISTS idx_events_time ON events(time);
	CREATE INDEX IF NOT EXISTS idx_campaign_targets_rid ON campaign_targets(rid);
	CREATE INDEX IF NOT EXISTS idx_questions_assessment ON questions(assessment_id);
	CREATE INDEX IF NOT EXISTS idx_answer_options_question ON answer_options(question_id);
	CREATE INDEX IF NOT EXISTS idx_user_attempts_user ON user_assessment_attempts(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_attempts_assessment ON user_assessment_attempts(assessment_id);
	CREATE INDEX IF NOT EXISTS idx_user_responses_attempt ON user_responses(attempt_id);
	`

	_, err := DB.Exec(schema)
	return err
}

func runMigrations() error {
	// Check if role column exists in users table
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='role'").Scan(&count)
	if err != nil {
		return err
	}

	// Add role column if it doesn't exist
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'")
		if err != nil {
			return err
		}
		// Update existing admin user to have admin role
		_, err = DB.Exec("UPDATE users SET role = 'admin' WHERE username = 'admin'")
		if err != nil {
			return err
		}
		log.Println("Migration: Added role column to users table")
	}

	// Migrate user API key column
	if err := MigrateUserAPIKey(); err != nil {
		return err
	}

	// Add email column to users table if it doesn't exist
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='email'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE users ADD COLUMN email TEXT")
		if err != nil {
			return err
		}
		log.Println("Migration: Added email column to users table")
	}

	// Add dynamic assessment columns to assessments table
	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('assessments') WHERE name='is_dynamic'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE assessments ADD COLUMN is_dynamic INTEGER DEFAULT 0")
		if err != nil {
			return err
		}
		log.Println("Migration: Added is_dynamic column to assessments table")
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('assessments') WHERE name='behavior_type'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE assessments ADD COLUMN behavior_type TEXT")
		if err != nil {
			return err
		}
		log.Println("Migration: Added behavior_type column to assessments table")
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('assessments') WHERE name='target_email'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE assessments ADD COLUMN target_email TEXT")
		if err != nil {
			return err
		}
		log.Println("Migration: Added target_email column to assessments table")
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('assessments') WHERE name='campaign_id'").Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = DB.Exec("ALTER TABLE assessments ADD COLUMN campaign_id INTEGER")
		if err != nil {
			return err
		}
		log.Println("Migration: Added campaign_id column to assessments table")
	}

	return nil
}

func CreateDefaultAdmin() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = DB.Exec("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
			"admin", string(hash), "admin", time.Now())
		if err != nil {
			return err
		}

		log.Println("Default admin user created (username: admin, password: admin)")
		log.Println("Please change the password after first login!")
	}

	return nil
}
