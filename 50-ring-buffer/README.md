# 50 · Fixed-size ring buffer

## Problem

The Tier 3 capstone: a FIFO queue over a **fixed** backing array, where the read and write positions chase each other around the circle and index arithmetic wraps with `%`. This is #44's queue with the memory problem *solved* — storage never grows, dead slots are reused — and it's real infrastructure: Go channels are ring buffers internally, as are log buffers, audio pipelines, and network windows. The classic design trap is built in: with only a head and a tail index, `head == tail` is ambiguous — it means both "empty" and "full". Resolving that ambiguity is the design decision at the center of the drill.

## Contract (what the tests enforce)

```go
type Ring struct {
    buf   []int
    head  int // next slot to read
    tail  int // next slot to write
    count int // number of stored elements — the ambiguity resolver
}

func NewRing(capacity int) (*Ring, error)  // error when capacity <= 0
func (r *Ring) Write(v int) bool           // false when full — value NOT stored
func (r *Ring) Read() (int, bool)          // false when empty
func (r *Ring) Len() int                   // elements currently stored
func (r *Ring) Cap() int                   // fixed capacity
```

- **Policy pinned: reject-when-full.** `Write` on a full ring returns `false` and stores nothing; existing data is never overwritten. (The alternative — overwrite-oldest, what log buffers do — is equally valid engineering but a different contract; the tests enforce this one.)
- **FIFO through the wrap:** values come out in write order even after the indices have lapped the array many times. `Write` stores at `tail` and advances it circularly; `Read` takes from `head` and advances it circularly; both wrap with `% capacity`.
- **The explicit `count` field is the pinned ambiguity fix.** Without it, empty and full are indistinguishable at `head == tail`. The alternatives (sacrificing one slot so full is `(tail+1)%cap == head`, or monotonic uncapped counters) are worth a comment; the count is the simplest and is what the struct above carries.
- **One allocation, ever:** the backing slice is created in `NewRing` and never grows, shrinks, or reallocates. No `append` anywhere.
- `NewRing(0)` and negative capacities → error, not a panic and not a useless ring.
- Zero and negative *values* are ordinary data (`Read`'s `(0, false)` on empty is disambiguated by the bool, as drilled in #43/#44).
- A drained ring is indistinguishable from a fresh one: fill-drain-fill cycles work forever.

## Worked examples

**Example 1 — the pinned scripted sequence (capacity 3):**

```go
r, _ := NewRing(3)
r.Write(1)   // true          [1 _ _]  len 1
r.Write(2)   // true          [1 2 _]  len 2
r.Write(3)   // true          [1 2 3]  len 3 — full
r.Write(4)   // FALSE         [1 2 3]  4 rejected, nothing lost
r.Read()     // (1, true)     [_ 2 3]  len 2 — a slot freed
r.Write(4)   // true          [4 2 3]  4 written into the RECYCLED slot 0
r.Read()     // (2, true)
r.Read()     // (3, true)
r.Read()     // (4, true)     — FIFO held straight through the wrap
r.Read()     // (0, false)    — empty
```

That `Write(4)` landing in physical slot 0 while logical order stays 2,3,4 *is* the ring buffer.

**Example 2 — why `count` exists:**

```
capacity 3, after Write,Write,Write:  head == 0, tail == 0  (tail wrapped) — FULL
capacity 3, after 3 Writes + 3 Reads: head == 0, tail == 0  — EMPTY
Same indices. Only count (3 vs 0) tells them apart.
```

**Example 3 — capacity 1, the smallest ring:**

```go
r, _ := NewRing(1)
r.Write(7)   // true
r.Write(8)   // false — full at one element
r.Read()     // (7, true)
r.Write(8)   // true — same slot, reused
r.Read()     // (8, true)
```

## Edge cases the tests cover

- `NewRing(0)` and `NewRing(-3)` → error.
- The full scripted sequence from Example 1, step by step with `Len` checks.
- Write to full (rejected, contents intact); read from empty; read-after-drain repeatedly false.
- Capacity 1 lifecycle.
- Fill → drain → fill again cycles (a used ring behaves like new).
- Zero and negative stored values.
- **The wrap stress:** 1,000 alternating write/read operations on a capacity-4 ring, asserting FIFO order throughout — the indices lap the array 250 times; any modulo slip fails within the first few laps.
- **The oracle:** 5,000 random operations (fixed seed) on a capacity-4 ring, mirrored against a trivially correct model queue in the test — every Write's accept/reject and every Read's value must match the model exactly. This is the test that leaves no corner for an off-by-one to hide in.
- `Len`/`Cap` consistency after every step everywhere.
