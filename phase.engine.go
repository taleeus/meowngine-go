package meowngine

import (
	"context"
	"fmt"
	"iter"
	"math"
)

// Phase is just an index.
//
// Phases will be dispatched in order.
// The World lifecycle is:
//   - ON_START (0)
//   - Game loop (all Phases from 1 to math.MaxUint excluded)
//   - ON_END (math.MaxUint)
//
// You can add your Phases in between the default ones: check the indexes.
type Phase uint

const (
	ON_START    Phase = 0
	PRE_FRAME   Phase = 100
	ON_LOAD     Phase = 200
	POST_LOAD   Phase = 300
	PRE_UPDATE  Phase = 400
	ON_UPDATE   Phase = 500
	ON_VALIDATE Phase = 600
	POST_UPDATE Phase = 700
	PRE_STORE   Phase = 800
	ON_STORE    Phase = 900
	POST_FRAME  Phase = 1000
	ON_END      Phase = math.MaxUint
)

// String returns the Phase name
func (phase Phase) String() string {
	switch {
	case phase == ON_END:
		return "OnEnd"
	case phase >= POST_FRAME:
		return fmt.Sprintf("PostFrame(%d)", phase-POST_FRAME)
	case phase >= ON_STORE:
		return fmt.Sprintf("OnStore(%d)", phase-ON_STORE)
	case phase >= PRE_STORE:
		return fmt.Sprintf("PreStore(%d)", phase-PRE_STORE)
	case phase >= POST_UPDATE:
		return fmt.Sprintf("PostUpdate(%d)", phase-POST_UPDATE)
	case phase >= ON_VALIDATE:
		return fmt.Sprintf("OnValidate(%d)", phase-ON_VALIDATE)
	case phase >= ON_UPDATE:
		return fmt.Sprintf("OnUpdate(%d)", phase-ON_UPDATE)
	case phase >= PRE_UPDATE:
		return fmt.Sprintf("PreUpdate(%d)", phase-PRE_UPDATE)
	case phase >= POST_LOAD:
		return fmt.Sprintf("PostLoad(%d)", phase-POST_LOAD)
	case phase >= ON_LOAD:
		return fmt.Sprintf("OnLoad(%d)", phase-ON_LOAD)
	case phase >= PRE_FRAME:
		return fmt.Sprintf("PreFrame(%d)", phase-PRE_FRAME)
	default:
		return fmt.Sprintf("OnStart(%d)", phase-ON_START)
	}
}

// System is where the game logic lives.
//
// Systems on the same Phase run automatically in parallel. If it's not the desired behaviour, give the operations an order
// using a different Phase.
type System[S any] func(context.Context, *World[S], iter.Seq[Entity[S]]) error

// Action is an operation that should run on a given Phase.
// This is different from a System for two reasons:
//   - Actions are used for platform and lifetime logic. They don't operate on Entities
//   - For the previous reason, Actions run on the main thread sequentially
type Action[S any] func(context.Context, *World[S]) error
