package command

import (
	"context"
	"errors"
)

var (
	ErrNotReceiver    = errors.New("not implement receiver interface")
	ErrNotCommand     = errors.New("not implement Command interface")
	ErrNotUndoCommand = errors.New("not implement UndoCommand interface")
	ErrNotRedoCommand = errors.New("not implement RedoCommand interface")
)

// A Command encapsulates a unit of processing work to be performed.
type Command interface {
	// Execute a unit of processing work to be performed
	Execute(ctx context.Context) (context.Context, error)
}

// UndoCommand undo a Command
type UndoCommand interface {
	Undo(ctx context.Context) (context.Context, error)
}

// RedoCommand redo a Command
type RedoCommand interface {
	Redo(ctx context.Context) (context.Context, error)
}

// The CommandFunc type is an adapter to allow the use of ordinary functions as Command.
// If f is a function with the appropriate signature, CommandFunc(f) is a Command that calls f.
type CommandFunc func(ctx context.Context) (context.Context, error)

// Execute calls f(ctx).
func (f CommandFunc) Execute(ctx context.Context) (context.Context, error) {
	return f(ctx)
}

// The UndoCommandFunc type is an adapter to allow the use of ordinary functions as UndoCommand.
// If f is a function with the appropriate signature, UndoCommandFunc(f) is a UndoCommand that calls f.
type UndoCommandFunc func(ctx context.Context) (context.Context, error)

// Undo calls f(ctx).
func (f UndoCommandFunc) Undo(ctx context.Context) (context.Context, error) {
	return f(ctx)
}

// The RedoCommandFunc type is an adapter to allow the use of ordinary functions as RedoCommand.
// If f is a function with the appropriate signature, RedoCommandFunc(f) is a RedoCommand that calls f.
type RedoCommandFunc func(ctx context.Context) (context.Context, error)

// Redo calls f(ctx).
func (f RedoCommandFunc) Redo(ctx context.Context) (context.Context, error) {
	return f(ctx)
}

type ConcreteCommand struct {
	receiver Receiver
}

func NewConcreteCommand(receiver Receiver) *ConcreteCommand {
	return &ConcreteCommand{receiver: receiver}
}

func (cmd *ConcreteCommand) Execute(ctx context.Context) error {
	if cmd.receiver == nil {
		return ErrNotReceiver
	}
	return cmd.receiver.Action(ctx)
}

type RichCommand struct {
	cmd     Command
	undoCmd UndoCommand
	redoCmd RedoCommand
}

func NewRichCommand(cmd Command, undoCmd UndoCommand, redoCmd RedoCommand) *RichCommand {
	return &RichCommand{cmd: cmd, undoCmd: undoCmd, redoCmd: redoCmd}
}

func (cmd *RichCommand) Execute(ctx context.Context) (context.Context, error) {
	if cmd.cmd == nil {
		return ctx, ErrNotCommand
	}
	return cmd.cmd.Execute(ctx)
}

func (cmd *RichCommand) Undo(ctx context.Context) (context.Context, error) {
	if cmd.undoCmd == nil {
		return ctx, ErrNotUndoCommand
	}
	return cmd.undoCmd.Undo(ctx)
}

func (cmd *RichCommand) Redo(ctx context.Context) (context.Context, error) {
	if cmd.redoCmd == nil {
		return ctx, ErrNotRedoCommand
	}
	return cmd.redoCmd.Redo(ctx)
}
