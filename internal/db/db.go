package db

import (
	"database/sql"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title TEXT NOT NULL,
	comment TEXT,
	repeat VARCHAR(128)
);

CREATE INDEX idx_date_scheduler ON scheduler(date);
`

var	db *sql.DB

var (
	ErrOpenDB = errors.New("не удалось открыть базу данных")
	ErrInitDB = errors.New("не удалось инициализировать базу данных")
)

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return ErrOpenDB
	}
	if !install {
		return nil
	}

	_, err = db.Exec(schema)
	if err != nil {
		return ErrInitDB
	}

	return nil
}

func Close() {
	db.Close()
}