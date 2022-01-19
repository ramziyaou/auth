package myerrors

import "errors"

var (
	ErrDuplicateUser   = errors.New("username or IIN exists")
	ErrInvalidAcc      = errors.New("invalid account(s)")
	ErrInvalidAmt      = errors.New("invalid amount")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidUsername = errors.New("invalid username")
	ErrNilTemplate     = errors.New("template not found")
	ErrRefreshNotFound = errors.New("refresh token not found in redis")
	ErrSameAccount     = errors.New("transfer between same account not allowed")
	ErrTokenExpired    = errors.New("Token is expired")
	ErrTokenMismatch   = errors.New("tokens don't match")
	ErrCookieNotFound  = errors.New("token cookie not found")
	ErrUserNotFound    = errors.New("user not found")
)
