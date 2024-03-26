package list

import "sync/atomic"

type node[T any] struct {
	value T
	next  atomic.Pointer[node[T]]
}
