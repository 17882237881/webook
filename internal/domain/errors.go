package domain

import "errors"

var (
	ErrInvalidUserOrPassword = errors.New("invalid user or password")
	ErrDuplicateEmail        = errors.New("duplicate email")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrUserNotFound          = errors.New("user not found")
	ErrPostNotFound          = errors.New("post not found")
	ErrPostNotAuthor         = errors.New("post not author")
	ErrPostAlreadyPublished  = errors.New("post already published")
)
