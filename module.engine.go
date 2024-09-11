package meowngine

// Module is a set of configurations, packed in the same reusable unit
type Module[S any] interface {
	Configure(*World[S])
}
