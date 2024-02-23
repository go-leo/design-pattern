package cqrs

import "errors"

var (
	// ErrHandlerNil CommandHandler or QueryHandler is nil
	ErrHandlerNil = errors.New("handler is nil")

	// ErrRegistered not register CommandHandler or QueryHandler
	ErrRegistered = errors.New("handler registered")

	// ErrUnregistered is not register CommandHandler or QueryHandler
	ErrUnregistered = errors.New("handler unregistered")

	// ErrCommandNil Command arg is nil
	ErrCommandNil = errors.New("command is nil")

	// ErrQueryNil Query arg is nil
	ErrQueryNil = errors.New("query is nil")

	// ErrUnimplemented handler is not implement CommandHandler or QueryHandler
	ErrUnimplemented = errors.New("handler is not implement CommandHandler or QueryHandler")

	// ErrBusClosed bus was closed
	ErrBusClosed = errors.New("bus was closed")
)
