package meowngine

import "errors"

var ErrNoEntity = errors.New("entity not found")
var ErrEntityDeleted = errors.New("entity is marked deleted")
var ErrComponentCleanupFailed = errors.New("entity deleted, but failed to remove some components")

var ErrAlreadyAssigned = errors.New("component is already assigned to entity")
var ErrNotAssigned = errors.New("component is not assigned to entity")
var ErrTypeAssertionFailed = errors.New("component type assertion failed")

var ErrRegistryBadSize = errors.New("registry size does not match expectations")

// ErrFatal is used to mark errors that should stop the application. Wrap or join it to the error that you want to mark
var ErrFatal = errors.New("fatal error")
