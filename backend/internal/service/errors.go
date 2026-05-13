package service

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUpdateFailed      = errors.New("update failed")
	ErrDeleteFailed      = errors.New("delete failed")
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskForbidden     = errors.New("task does not belong to user")
)
