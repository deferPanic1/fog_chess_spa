package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolation &&
			strings.Contains(pgErr.ConstraintName, constraintName)
	}
	return false
}
