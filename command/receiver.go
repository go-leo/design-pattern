package command

import "context"

type Receiver interface {
	Action(ctx context.Context) error
}
