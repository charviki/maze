package service

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists indicates a resource with the same key already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrInvalidInput indicates the request has invalid or missing fields.
	ErrInvalidInput = errors.New("invalid input")
)
