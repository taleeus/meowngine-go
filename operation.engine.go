package meowngine

import (
	"context"
	"errors"
	"iter"

	"github.com/taleeus/commons-go/itex"
)

// Operation is a callback that operates on a piece of data given by a PipedSystem
type Operation[S, D any] func(context.Context, *World[S], D) error

// PipedSystem builds a System, splitting the piping part from the consuming one.
//
// Use this if you prefer to isolate game logic in a simple function (the Operation)
// and if you have Pipes that are common between various Systems.
func PipedSystem[S, D any](sysPipe itex.PipeFn[Entity[S], D], op Operation[S, D]) System[S] {
	return func(ctx context.Context, world *World[S], entSeq iter.Seq[Entity[S]]) error {
		var err error
		for data := range itex.Pipe(entSeq, sysPipe) {
			err = errors.Join(err, op(ctx, world, data))
		}

		return err
	}
}
