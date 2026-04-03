package contract

// Resolver defines a generic interface for resolving implementations by name.
type Resolver[T any] interface {
	Resolve(name string) (T, error)
}

// ResolverFunc adapts a function to the Resolver interface.
type ResolverFunc[T any] func(name string) (T, error)

// Resolve delegates to the wrapped function.
func (r ResolverFunc[T]) Resolve(name string) (T, error) {
	return r(name)
}
