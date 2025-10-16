package services

import (
	"encoding/json"
	"errors"
)

var (
	ErrInternal = errors.New("internal error")
)

func FromJSON[T any](data []byte) (T, error) {
	o := new(T)
	err := json.Unmarshal(data, &o)
	return *o, err
}
