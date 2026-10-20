package common

import (
	"database/sql"
	"fmt"
)

// InTx runs fn inside a transaction: it commits when fn succeeds and rolls
// back when fn (or the commit) fails.
func InTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction. %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction. %w", err)
	}

	return nil
}
