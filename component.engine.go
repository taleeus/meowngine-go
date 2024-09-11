package meowngine

import (
	"errors"
	"fmt"
)

// Component is the base interface for all data used by Entities and Systems.
// Entities can have only one Component for each Component implementation
type Component interface {
	Name() string
}

type componentWrapper[C Component] struct {
	assigned bool
	Data     C
}

// GetComponent searches a Component of the provided type in the given entity
func GetComponent[C Component, S any](ent Entity[S]) (*C, error) {
	if ent.deleted {
		return nil, ErrEntityDeleted
	}

	var keyComp C
	reg, ok := getRegistry[C](ent.world)
	if !ok {
		return nil, fmt.Errorf("registry for %s not present (%w)", keyComp.Name(), ErrNotAssigned)
	}

	if ent.id >= ID(len(reg.items)) {
		return nil, fmt.Errorf("bad size in registry for %s (%w)", keyComp.Name(), ErrRegistryBadSize)
	}

	if !reg.items[ent.id].assigned {
		return nil, ErrNotAssigned
	}

	return &reg.items[ent.id].Data, nil
}

// HasComponent checks if the given Entity has the provided Component
func HasComponent[C Component, S any](ent Entity[S]) (bool, error) {
	comp, err := GetComponent[C](ent)
	if err != nil && !errors.Is(err, ErrNotAssigned) {
		return false, err
	}

	return comp != nil, nil
}

// SetComponent associates a Component to the given Entity
func SetComponent[C Component, S any](ent Entity[S], comp C) error {
	if ent.deleted {
		return ErrEntityDeleted
	}

	reg, ok := getRegistry[C](ent.world)
	if !ok {
		reg = assignRegistry[C](ent.world)
	}

	reg.assign(ent.id, comp)
	return nil
}
