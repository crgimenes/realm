package sqlite

import (
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	DB *sqlx.DB
)

func New(filename string) error {
	var err error

	databaseName := strings.Join([]string{
		"file:",
		filename,
		"?cache=shared&mode=rwc"}, "")

	DB, err = sqlx.Connect("sqlite", databaseName)

	return err
}
