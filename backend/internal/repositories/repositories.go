package repositories

import "github.com/go-jet/jet/v2/postgres"

func psqlStr(s string) postgres.StringExpression {
	return postgres.String(s)
}

type scannable interface {
	Scan(...any) error
}

type ErrInvalidField struct {
	Field string
}

func (e ErrInvalidField) Error() string {
	return "invalid field: " + e.Field
}
