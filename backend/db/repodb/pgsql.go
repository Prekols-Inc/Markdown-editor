package repodb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type DBConnAttrs struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}

type PGSQLRepo struct {
	db *sql.DB
}

func NewPGSQLRepo(attrs DBConnAttrs) (*PGSQLRepo, error) {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", attrs.host, attrs.port, attrs.user, attrs.password, attrs.dbname, attrs.sslmode),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	query := `
		CREATE TABLE IF NOT EXISTS documents (
    		title VARCHAR(255) PRIMARY KEY,
    		content BYTEA NOT NULL,
    		created_at TIMESTAMP DEFAULT NOW(),
    		updated_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &PGSQLRepo{db: db}, nil
}

func (r *PGSQLRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func convertError(err error) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return ErrFileNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			return ErrFileExists
		}
	}
	return err
}

func (r *PGSQLRepo) Save(filename string, data []byte) error {
	query := `
		UPDATE documents 
		SET content = $1, updated_at = $2 
		WHERE title = $3
	`
	result, err := r.db.Exec(query, data, time.Now(), filename)
	if err != nil {
		return convertError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrFileNotFound
	}

	return convertError(err)
}

func (r *PGSQLRepo) Create(filename string, data []byte) error {
	query := `
		INSERT INTO documents (title, content, updated_at) 
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, filename, data, time.Now())

	return convertError(err)
}

func (r *PGSQLRepo) Get(filename string) ([]byte, error) {
	var content []byte
	query := "SELECT content FROM documents WHERE title = $1 limit 1"
	err := r.db.QueryRow(query, filename).Scan(&content)
	if err != nil {
		return nil, convertError(err)
	}

	return content, nil
}

func (r *PGSQLRepo) Delete(filename string) error {
	query := "DELETE FROM documents WHERE title = $1;"
	result, err := r.db.Exec(query, filename)
	if err != nil {
		return convertError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrFileNotFound
	}

	return convertError(err)
}

func (r *PGSQLRepo) GetList() ([]string, error) {
	var filenames []string
	query := "SELECT title FROM documents"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, convertError(err)
		}
		filenames = append(filenames, title)
	}

	if err := rows.Err(); err != nil {
		return nil, convertError(err)
	}

	return filenames, nil
}
