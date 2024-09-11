package meowngine

import (
	"fmt"
)

const REGISTRY_INIT_CAP = 128

type registry[C Component] struct {
	items []componentWrapper[C]
}

func newRegistry[C Component](lastID ID) *registry[C] {
	multiplier := lastID/REGISTRY_INIT_CAP + 1

	var registry registry[C]
	registry.items = make([]componentWrapper[C], lastID+1, multiplier*REGISTRY_INIT_CAP)

	return &registry
}

func (reg *registry[C]) increment() {
	reg.items = append(reg.items, componentWrapper[C]{})
}

func (reg *registry[C]) assign(id ID, data C) (*C, error) {
	if id+1 > ID(len(reg.items)) {
		return nil, fmt.Errorf("%w (%d)", ErrNoEntity, id)
	}

	compWrap := &reg.items[id]
	if compWrap.assigned {
		return nil, fmt.Errorf("%w (%d, %s)", ErrAlreadyAssigned, id, compWrap.Data.Name())
	}

	compWrap.assigned = true
	compWrap.Data = data

	return &compWrap.Data, nil
}

func (reg *registry[C]) remove(id ID) error {
	if id+1 > ID(len(reg.items)) {
		return fmt.Errorf("%w (%d)", ErrNoEntity, id)
	}

	compWrap := &reg.items[id]
	if !compWrap.assigned {
		return fmt.Errorf("%w (%d, %s)", ErrNotAssigned, id, compWrap.Data.Name())
	}

	compWrap.assigned = false
	return nil
}
