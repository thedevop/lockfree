package list

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue[int]()
	require.NotNil(t, q)
	require.IsType(t, &Queue[int]{}, q)
}

func TestQueueEnqueue(t *testing.T) {
	v := "test"
	q := NewQueue[string]()
	q.Enqueue(v)
	require.Equal(t, v, q.head.Load().next.Load().value)
}

func TestQueueDequeue(t *testing.T) {
	v := "test"
	q := NewQueue[string]()
	q.Enqueue(v)
	r, ok := q.Dequeue()
	require.True(t, ok)
	require.Equal(t, v, r)
}

func TestQueue(t *testing.T) {
	const Threads = 100
	const Count = 10000

	results := testQueue(Threads, Threads, Count, false)

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

func testQueue(wthreads, rthreads, count int, sequential bool) []map[int]bool {
	q := NewQueue[int]()

	ctx, cancel := context.WithCancel(context.Background())

	results := make([]map[int]bool, rthreads)
	enq := sync.WaitGroup{}
	enq.Add(rthreads)
	for i := 0; i < rthreads; i++ {
		result := make(map[int]bool)
		results[i] = result
		go func(i int) {
			defer enq.Done()
			if sequential {
				<-ctx.Done()
			}
			for {
				v, ok := q.Dequeue()
				if ok {
					result[v] = true
				}
				if ctx.Err() != nil && q.Len() == 0 {
					break
				}
			}
		}(i)
	}

	deq := sync.WaitGroup{}
	deq.Add(wthreads)
	for i := 0; i < wthreads; i++ {
		go func(i int) {
			defer deq.Done()
			start := i * count
			end := start + count
			for n := start; n < end; n++ {
				q.Enqueue(n)
			}
		}(i)
	}

	deq.Wait()
	cancel()
	enq.Wait()

	return results
}

func BenchmarkQueue(b *testing.B) {
	const Threads = 100
	const Count = 10000

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testQueue(Threads, Threads, Count, false)
	}
}

func BenchmarkQueueSeq(b *testing.B) {
	const Threads = 100
	const Count = 10000

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testQueue(Threads, Threads, Count, true)
	}
}
