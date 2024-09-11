package meowngine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"unsafe"

	"golang.org/x/sync/errgroup"

	"github.com/taleeus/commons-go/data"
	"github.com/taleeus/commons-go/itex"
)

// World coordinates the game lifecycle, keeps track of Components and
// manages the interactions between Entities, Systems and Actions.
//
// To create a World, you MUST initalize it calling `engine.NewWorld()`.
// After you configured the game, call World.Run() to launch the game.
type World[S any] struct {
	State S

	entities   []Entity[S]
	deletedIDs []ID

	registries map[string]*registry[Component]

	phases  []Phase
	actions map[Phase][]Action[S]
	systems map[Phase][]System[S]

	shouldQuit bool
}

// NewWorld initializes all World internal data and returns it
func NewWorld[S any](state S) *World[S] {
	slog.Info("Creating a world", "state", state)

	world := World[S]{State: state}
	world.registries = make(map[string]*registry[Component])

	world.systems = make(map[Phase][]System[S])
	world.actions = make(map[Phase][]Action[S])

	return &world
}

// SpawnEntity creates a new Entity and returns it
func (world *World[S]) SpawnEntity(ctx context.Context) Entity[S] {
	var id ID
	switch len(world.deletedIDs) > 0 {
	case true:
		id = world.deletedIDs[0]
		world.deletedIDs = slices.Delete(world.deletedIDs, 0, 1)

		world.entities[id] = Entity[S]{
			id:    id,
			world: world,
			tags:  make(data.Set[string]),
		}

	case false:
		id = ID(len(world.entities))
		for _, registry := range world.registries {
			registry.increment()
		}

		world.entities = append(world.entities, Entity[S]{
			id:    id,
			world: world,
			tags:  make(data.Set[string]),
		})
	}

	slog.InfoContext(ctx, "Spawning entity", "id", id)
	return world.entities[id]
}

// DeleteEntity removes an Entity from the World
func (world *World[S]) DeleteEntity(ctx context.Context, id ID) error {
	if id+1 > ID(len(world.entities)) {
		return fmt.Errorf("%w (%d)", ErrNoEntity, id)
	}

	entity := world.entities[id]
	if entity.deleted {
		return fmt.Errorf("%w (%d)", ErrEntityDeleted, id)
	}

	entity.deleted = true
	world.entities[id] = entity
	world.deletedIDs = append(world.deletedIDs, id)

	var regerr error
	for _, registry := range world.registries {
		regerr = errors.Join(regerr, registry.remove(id))
	}

	if regerr != nil {
		return fmt.Errorf("%w\n%w", ErrComponentCleanupFailed, regerr)
	}

	slog.InfoContext(ctx, "Deleting entity", "id", id)
	return nil
}

func getRegistry[C Component, S any](world *World[S]) (*registry[C], bool) {
	var keyComp C
	reg, ok := world.registries[keyComp.Name()]
	if !ok {
		return nil, false
	}

	unsafeReg := (*registry[C])(unsafe.Pointer(reg))
	return unsafeReg, true
}

func assignRegistry[C Component, S any](world *World[S]) *registry[C] {
	reg := newRegistry[C](ID(len(world.entities)))

	var keyComp C
	unsafeReg := (*registry[Component])(unsafe.Pointer(reg))

	world.registries[keyComp.Name()] = unsafeReg
	return reg
}

// System registers a System in the given Phase
func (world *World[S]) System(phase Phase, system System[S]) *World[S] {
	world.registerPhase(phase)
	world.systems[phase] = append(world.systems[phase], system)

	return world
}

// Action registers a Action in the given Phase
func (world *World[S]) Action(phase Phase, action Action[S]) *World[S] {
	world.registerPhase(phase)
	world.actions[phase] = append(world.actions[phase], action)

	return world
}

// Use installs a Module into the World
func (world *World[S]) Use(module Module[S]) *World[S] {
	module.Configure(world)
	return world
}

func (world *World[S]) registerPhase(phase Phase) {
	if !slices.Contains(world.phases, phase) {
		world.phases = append(world.phases, phase)
		slices.Sort(world.phases)
	}
}

// Run launches the game. Check the Phase documentation to learn the World lifecycle
func (world *World[S]) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting world")
	if err := world.dispatchPhase(ctx, ON_START); err != nil && errors.Is(err, ErrFatal) {
		slog.ErrorContext(ctx, "Fatal error during OnStart; quitting")
		return err
	}

	loopPhases := itex.Apply(world.phases,
		itex.Filter(func(p Phase) bool { return p != ON_START && p != ON_END }),
	)
	slog.InfoContext(ctx, "Starting game loop",
		"loopPhases", itex.Apply(loopPhases,
			itex.Map(func(p Phase) string { return p.String() }),
		),
	)
	for !world.shouldQuit {
		for _, phase := range loopPhases {
			if err := world.dispatchPhase(ctx, phase); err != nil && errors.Is(err, ErrFatal) {
				slog.ErrorContext(ctx, "Fatal error found; quitting")
				return err
			}
		}
	}

	slog.InfoContext(ctx, "Ending world")
	if err := world.dispatchPhase(ctx, ON_END); err != nil && errors.Is(err, ErrFatal) {
		slog.ErrorContext(ctx, "Fatal error during OnStart; quitting")
		return err
	}

	return nil
}

// Quit signals the World that it should quit when the current gameloop ends
func (world *World[S]) Quit(ctx context.Context) {
	slog.InfoContext(ctx, "Quit signaled")
	world.shouldQuit = true
}

func (world *World[S]) dispatchPhase(ctx context.Context, phase Phase) error {
	var actionErr error
	for _, action := range world.actions[phase] {
		actionErr = errors.Join(actionErr, action(ctx, world))
	}

	if actionErr != nil {
		slog.WarnContext(ctx, "Error during phase actions",
			"phase", phase.String(),
			"err", actionErr.Error(),
		)

		return actionErr
	}

	eg, phaseCtx := errgroup.WithContext(ctx)
	for _, sys := range world.systems[phase] {
		entIter := itex.Pipe(slices.Values(world.entities), itex.Filter(func(e Entity[S]) bool {
			return !e.deleted
		}))

		eg.Go(func() error {
			return sys(phaseCtx, world, entIter)
		})
	}

	if err := eg.Wait(); err != nil {
		slog.WarnContext(ctx, "Error during phase",
			"phase", phase.String(),
			"err", err.Error(),
		)

		return err
	}

	return nil
}
