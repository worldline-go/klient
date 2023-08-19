package klient

type Null[T any] struct {
	Value T
	Valid bool
}
