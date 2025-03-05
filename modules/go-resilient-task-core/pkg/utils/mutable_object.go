package u

type MutableObject[T any] struct {
	Value T
}

// NewMutableObject creates a new MutableObject with the given value
func NewMutableObject[T any](value T) *MutableObject[T] {
	return &MutableObject[T]{Value: value}
}

// Set updates the value of the MutableObject
func (m *MutableObject[T]) Set(value T) {
	m.Value = value
}

// Get returns the current value of the MutableObject
func (m *MutableObject[T]) Get() T {
	return m.Value
}
