package meowngine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/taleeus/commons-go/itex"
)

type testComp struct {
	msg string
}

func (testComp) Name() string {
	return "testComp"
}

func TestEntity(t *testing.T) {
	ctx := context.Background()
	world := NewWorld(&struct{}{})

	ent := world.SpawnEntity(ctx)
	SetComponent(ent, testComp{msg: "hello"})

	comp, err := GetComponent[testComp](ent)
	if err != nil {
		t.Fatal(err)
	}

	if comp.msg != "hello" {
		t.Fatal("wrong component data")
	}

	oldID := ent.id
	world.DeleteEntity(ctx, ent.id)
	ent = world.SpawnEntity(ctx)

	if ent.id != oldID {
		t.Fatal("old ID has not been recycled")
	}

	if _, err := GetComponent[testComp](ent); !errors.Is(err, ErrNotAssigned) {
		t.Fatal("old component is still assigned")
	}

	newEnt := world.SpawnEntity(ctx)
	if newEnt.id <= ent.id {
		t.Fatal("new entity didn't increment ID")
	}
}

type testPrototype struct {
	Entity[any]
	comp *testComp
}

func testPipe() itex.PipeFn[Entity[any], testPrototype] {
	return itex.MapMaybe(func(ent Entity[any]) (testPrototype, bool) {
		comp, err := GetComponent[testComp](ent)
		if err != nil {
			return testPrototype{}, false
		}

		return testPrototype{
			Entity: ent,
			comp:   comp,
		}, true
	})
}

func testOperation(ctx context.Context, world *World[any], proto testPrototype) error {
	if proto.comp.msg != "hello" {
		return errors.Join(
			ErrFatal,
			fmt.Errorf("entity is not greeting me :( (msg: %s)", proto.comp.msg),
		)
	}

	return nil
}

func timerShutdownAction(ctx context.Context, world *World[any]) error {
	go func() {
		time.Sleep(time.Second * 3)
		world.Quit(ctx)
	}()

	return nil
}

type notifyShutdownAction struct {
	quitted *bool
}

func (nsa notifyShutdownAction) process(ctx context.Context, world *World[any]) error {
	*nsa.quitted = true
	return nil
}

func TestLifecycle(t *testing.T) {
	ctx := context.Background()
	world := NewWorld(any(struct{}{}))

	world.Action(ON_START, timerShutdownAction)

	ent := world.SpawnEntity(ctx)
	SetComponent(ent, testComp{msg: "hello"})

	world.System(ON_UPDATE, PipedSystem(testPipe(), testOperation))

	var quitted bool
	nsa := notifyShutdownAction{quitted: &quitted}
	world.Action(ON_END, nsa.process)

	go world.Run(ctx)

	time.Sleep(time.Second * 5)
	if !quitted {
		t.Fatal("world is still running")
	}
}
