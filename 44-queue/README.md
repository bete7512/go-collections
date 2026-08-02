# 44 · Queue

## Problem

The FIFO (first-in-first-out) counterpart to #43: enqueue at the back, dequeue from the front. The API is a mirror of the stack's — same `(value, ok)` convention, same zero-value usability — but this container carries a hidden cost the stack doesn't have, and *naming that cost* is part of the deliverable. Queues are the other half of everything: job queues, BFS traversals, message buffers, the channel semantics you'll live in throughout Tier 5.

## Contract (what the tests enforce)

```go
type Queue struct {
    items []int
}

func (q *Queue) Enqueue(v int)
func (q *Queue) Dequeue() (int, bool)
func (q *Queue) Len() int
func (q *Queue) Empty() bool
```

- **FIFO:** enqueue 1, 2, 3 → dequeue yields 1, then 2, then 3.
- **`Dequeue` on empty → `(0, false)`** — same convention as #43, same reasoning, and the tests enqueue `0` and `-1` to keep sentinels impossible.
- **Zero value usable:** `var q Queue` works without a constructor.
- Pointer receivers, unexported field — as before.
- **The required comment — the memory lesson:** the natural implementation dequeues with `q.items = q.items[1:]`. That's correct, and the tests pass with it — but it advances the slice pointer without ever freeing the front: the backing array retains every dequeued element, and `append` keeps growing it. A queue holding 10 live elements that has processed a million holds memory for… more than 10. Write a comment on `Dequeue` naming the problem and the three standard fixes:
  1. **Copy-down** occasionally (`copy(q.items, q.items[head:])`) to reclaim space;
  2. **Explicit head index** into a reused array, compacting when the dead prefix grows;
  3. **Ring buffer** — fixed storage, wrapped indices — which is exactly capstone #50.
  You don't have to implement them here; you have to know why they exist. (Channels, for comparison, are ring buffers internally.)

## Worked examples

**Example 1 — FIFO order:**

```go
var q Queue
q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)
q.Dequeue()   // → (1, true)   first in, first out
q.Dequeue()   // → (2, true)
q.Dequeue()   // → (3, true)
q.Dequeue()   // → (0, false)
```

**Example 2 — interleaving keeps order within what's present:**

```go
var q Queue
q.Enqueue(1)
q.Enqueue(2)
q.Dequeue()   // → (1, true)
q.Enqueue(3)
q.Dequeue()   // → (2, true)  — 2 was ahead of 3
q.Dequeue()   // → (3, true)
```

**Example 3 — stack vs queue in one picture:**

```
push/enqueue: 1 2 3
stack pops:   3 2 1   (LIFO — reversal)
queue deqs:   1 2 3   (FIFO — preservation)
```

Same slice underneath; which *end* you take from defines the structure.

## Edge cases the tests cover

- Zero-value queue: all four methods safe before any enqueue.
- Dequeue on empty; queue reusable after draining.
- Zero and negative values enqueued (sentinel trap).
- Duplicates.
- Long interleaved sequences (25+ operations) asserting FIFO order throughout.
- Len/Empty consistency after every operation in a scripted table.
- 10,000 enqueues dequeued in exact insertion order, ending Empty.
- Alternating enqueue/dequeue at length 1 — the pattern that makes the `[1:]` memory issue vivid (and still must work correctly).
