package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/cugu/fomo/db/sqlc"
)

var _ scs.Store = &SQLiteStore{}

// SQLiteStore is a session store that uses SQLite to persist session data.
// It implements the scs.Store interface.
type SQLiteStore struct {
	DB *sqlc.Queries
}

// Delete removes a session token and corresponding data from the store.
func (s *SQLiteStore) Delete(token string) error {
	return s.DB.DeleteSession(context.Background(), token)
}

// Find returns the data for a session token from the store.
func (s *SQLiteStore) Find(token string) (b []byte, found bool, err error) {
	session, err := s.DB.FindSession(context.Background(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("failed to find session: %w", err)
	}

	return session.Data, true, nil
}

// Commit adds the session token and data to the store, with the given expiry time
// if the session token already exists, then the data and expiry time are overwritten.
func (s *SQLiteStore) Commit(token string, b []byte, expiry time.Time) error {
	return s.DB.CommitSession(context.Background(), sqlc.CommitSessionParams{
		Token:  token,
		Data:   b,
		Expiry: expiry,
	})
}
