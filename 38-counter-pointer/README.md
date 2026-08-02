# 38 · Counter with a pointer receiver

## Problem

The companion piece to #37, and the single most important mechanical lesson in Go methods: **a value receiver gets a copy, so mutations vanish; a pointer receiver gets the original.** Rather than telling you to avoid the bug, this drill makes you *ship* it: you'll write both the broken value-receiver increment and the correct pointer-receiver one, and the tests assert both behaviors — the broken one must stay broken. Feeling the copy disappear is how the rule sticks.

## Contract (what the tests enforce)

```go
type Counter struct{ n int }

func (c *Counter) Inc()          // pointer receiver — the mutation persists
func (c Counter) IncBroken()     // value receiver — mutates a copy, deliberately kept
func (c Counter) Value() int
```

- **`Inc` uses a pointer receiver.** Three calls → `Value()` returns 3.
- **`IncBroken` uses a value receiver and is part of the deliverable.** Its body increments exactly like `Inc` — same statement — but because the receiver is a copy, the caller's counter never changes. The tests call it three times and assert the counter is **still 0**. It exists as an executable demonstration of the bug, pinned forever.
- **The zero value is ready to use.** `var c Counter` needs no constructor — `Value()` is 0, `Inc()` works immediately. (This is a Go design virtue worth noticing: `bytes.Buffer`, `sync.Mutex`, and `strings.Builder` all work this way.)
- `Value()` reads without mutating.
- The field stays unexported — callers interact only through methods.
- **Why the call sites look identical:** `c.Inc()` on an addressable variable is shorthand for `(&c).Inc()` — Go takes the address for you. That's why the bug is invisible at the call site and lives only in the receiver declaration. Corollaries the tests exercise: elements of a slice are addressable (`s[0].Inc()` compiles and works); a counter reached through a pointer shares state with the original.
- One consequence you can only see at compile time (try it, then leave it commented out): **map elements are not addressable** — `m["a"].Inc()` does not compile. The fix is storing `*Counter` in the map.

## Worked examples

**Example 1 — the correct one:**

```go
var c Counter
c.Inc()
c.Inc()
c.Inc()
c.Value()   // → 3
```

**Example 2 — the bug, preserved:**

```go
var c Counter
c.IncBroken()   // increments a COPY of c; the copy is discarded
c.IncBroken()
c.IncBroken()
c.Value()   // → 0. Still. That's the lesson.
```

**Example 3 — pointers share, values don't:**

```go
a := Counter{}
p := &a
p.Inc()
a.Value()   // → 1: p and a are the same counter

b := a      // b is a copy taken at value 1
b.Inc()
a.Value()   // → still 1: b's increment is b's alone
```

## Edge cases the tests cover

- Zero value usable without a constructor.
- Three `Inc` → 3; a thousand `Inc` → 1000.
- Three `IncBroken` → still 0 (the pinned bug).
- Interleaved `Inc`/`IncBroken` — only the `Inc` calls count.
- Two counters mutate independently.
- A counter inside a slice: `s[0].Inc()` mutates the element in place.
- A counter reached through a pointer shares state; a copied counter diverges from the moment of the copy.
