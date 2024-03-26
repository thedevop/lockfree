package list

import (
	"container/list"
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStack(t *testing.T) {
	s := NewStack[int]()
	require.NotNil(t, s)
	require.IsType(t, &Stack[int]{}, s)
}

func TestStackPush(t *testing.T) {
	v := "test"
	s := NewStack[string]()
	s.Push(v)
	require.Equal(t, v, s.head.Load().value)
}

func TestStackPop(t *testing.T) {
	v := "test"
	s := NewStack[string]()
	s.Push(v)
	r, ok := s.Pop()
	require.True(t, ok)
	require.Equal(t, v, r)
}

func TestStack(t *testing.T) {
	const Threads = 20
	const Count = 10000

	results := testStack(Threads, Threads, Count, false)

	combinedResults := make(map[int]bool)
	for i := 0; i < Threads; i++ {
		for k, v := range results[i] {
			combinedResults[k] = v
		}
	}

	require.Equal(t, Threads*Count, len(combinedResults))

	for i := 0; i < Threads; i++ {
		start := i * Count
		end := start + Count
		for n := start; n < end; n++ {
			_, ok := combinedResults[n]
			require.True(t, ok)
		}
	}
}

func testStack(wthreads, rthreads, count int, sequential bool) []map[int]bool {
	s := NewStack[int]()

	ctx, cancel := context.WithCancel(context.Background())

	results := make([]map[int]bool, rthreads)
	pop := sync.WaitGroup{}
	pop.Add(rthreads)
	for i := 0; i < rthreads; i++ {
		result := make(map[int]bool)
		results[i] = result
		go func(i int) {
			defer pop.Done()
			if sequential {
				<-ctx.Done()
			}
			for {
				v, ok := s.Pop()
				if ok {
					result[v] = true
				}
				if ctx.Err() != nil && s.Len() == 0 {
					break
				}
			}
		}(i)
	}

	push := sync.WaitGroup{}
	push.Add(wthreads)
	for i := 0; i < wthreads; i++ {
		go func(i int) {
			defer push.Done()
			start := i * count
			end := start + count
			for n := start; n < end; n++ {
				s.Push(n)
			}
		}(i)
	}

	push.Wait()
	cancel()
	pop.Wait()

	return results
}

func BenchmarkStack(b *testing.B) {
	const Threads = 100
	const Count = 10000

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testStack(Threads, Threads, Count, false)
	}
}

func BenchmarkStackSeq(b *testing.B) {
	const Threads = 100
	const Count = 10000

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testStack(Threads, Threads, Count, true)
	}
}

func BenchmarkContainerList(b *testing.B) {
	const Threads = 100
	const Count = 10000

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testContainerList(Threads, Threads, Count, false)
	}
}

func testContainerList(wthreads, rthreads, count int, sequential bool) []map[int]bool {
	l := list.New()
	mutex := sync.Mutex{}

	ctx, cancel := context.WithCancel(context.Background())

	results := make([]map[int]bool, rthreads)
	pop := sync.WaitGroup{}
	pop.Add(rthreads)
	for i := 0; i < rthreads; i++ {
		result := make(map[int]bool)
		results[i] = result
		go func(i int) {
			defer pop.Done()
			if sequential {
				<-ctx.Done()
			}
			for {
				mutex.Lock()
				front := l.Front()
				if front != nil {
					l.Remove(front)
					mutex.Unlock()
					v := front.Value.(int)
					result[v] = true

				} else {
					mutex.Unlock()
				}
				if ctx.Err() != nil && l.Len() == 0 {
					break
				}
			}
		}(i)
	}

	push := sync.WaitGroup{}
	push.Add(wthreads)
	for i := 0; i < wthreads; i++ {
		go func(i int) {
			defer push.Done()
			start := i * count
			end := start + count
			for n := start; n < end; n++ {
				mutex.Lock()
				l.PushFront(n)
				mutex.Unlock()
			}
		}(i)
	}

	push.Wait()
	cancel()
	pop.Wait()

	return results
}
