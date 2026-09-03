# 80 · sync.Once

## The idea in one paragraph

Lazy initialization under concurrency has two failure modes: the work runs **more than once** (ten goroutines all see `cfg == nil` and all load it), or a caller reads the value **before initialization finished**. `sync.Once` kills both with one method: `once.Do(f)` runs `f` exactly once per `Once` value, and — the half people forget — **blocks every caller until that first invocation returns**. It is not fire-and-forget: when your `Do` comes back, the value is fully built, no matter which goroutine built it. This is the mechanism behind lazy singletons everywhere: DB connection pools, config loading, expensive regex compilation, feature-flag clients.

## Vocabulary

| term | meaning |
|---|---|
| `sync.Once` | zero-value-ready; guarantees its `Do` runs a function at most once, ever |
| `once.Do(f)` | runs `f` if nothing has run yet; otherwise waits (if the first run is in flight) or returns immediately |
| once per **Once value** | the guarantee belongs to the `Once`, not the function — `once.Do(g)` after `once.Do(f)` runs *nothing* |
| double-checked locking | the naive `if cfg == nil { cfg = load() }` idiom — a data race in Go, full stop |
| `sync.OnceFunc` / `sync.OnceValue` / `sync.OnceValues` | Go 1.21+ wrappers: `var getCfg = sync.OnceValue(load)` gives the same guarantee with no visible `Once`. Use raw `Once` for this drill; reach for `OnceValue` in real code |

Two sharp edges the spec pins:

- **A panic counts as done.** If `f` panics inside `Do`, the panic propagates *and the Once is spent* — no retry, ever; later callers get whatever half-state exists (usually nil). That's why init funcs must not panic on transient failures (a flaky network dial does not belong in one).
- **A used `Once` must not be copied** — `go vet` copylocks, same as #79's mutex.

---

## The API

| item | what it teaches |
|---|---|
| `Get` / `InitCalls` | the spec's drill: package-level lazy singleton |
| `LazyConfig` | the struct-field form — fresh instances make the guarantee testable repeatedly |
| `SameOnceTwice` | once per Once *value*, not per function |
| `RacyGet` | the naive nil-check, kept on display for `-race` to flag |

---

## Part 1 · package-level: `Get() *Config` and `InitCalls() int32`

```go
type Config struct{ /* fields are yours; the tests treat *Config as opaque */ }

func Get() *Config        // lazy-loads via a package-level sync.Once
func InitCalls() int32    // how many times the init function has ever run
```

**Pinned contract:**

- `Get` never returns nil, and every call — concurrent or sequential, first or thousandth — returns the **same pointer**.
- The init function increments a counter (make it atomic); `InitCalls()` reports it, and it must be exactly **1** after any number of `Get` calls. The tests hammer `Get` from 10 goroutines, then keep calling it sequentially, and check `InitCalls()` both times.
- Package-level state means the whole test binary shares one Once — which is exactly the point of this part.

**Traced example:**

```
10 goroutines call Get() simultaneously

  goroutine 3 wins: Do runs the init func   InitCalls → 1
  goroutines 0,1,2,4..9: Do BLOCKS until the init returns
  all 10 unblock → all 10 hold the identical *Config

later, main goroutine: Get() → same pointer, InitCalls() still 1
```

---

## Part 2 · struct-field form: `LazyConfig`

```go
func NewLazyConfig(load func() *Config) *LazyConfig
func (l *LazyConfig) Get() *Config
```

Same guarantee, but the `Once` lives in the struct and the loader is injected — so the tests can build a fresh instance per run and repeat the concurrency experiment twenty times.

**Pinned contract:**

- The injected `load` runs **exactly once** per `LazyConfig`, no matter how many goroutines call `Get`; every caller gets `load`'s return value, pointer-identical.
- **Blocking:** the tests inject a slow loader (tens of ms). No concurrent caller may ever observe nil — proof that `Do` waits rather than letting racers through.
- **Panic-marks-done:** if `load` panics, the first `Get` panics; every later `Get` returns nil **without panicking and without retrying** — the loader is never called again. The tests assert the loader ran exactly once through that whole sequence.

---

## Part 3 · `SameOnceTwice(first, second func())`

One fresh `sync.Once`; call `Do(first)` then `Do(second)` on it, sequentially, and return. **Pinned contract:** `first` executes, `second` never does — the tests pass closures that flip test-side flags. This is the once-per-Once-value demonstration: the second `Do` doesn't run a "different function", it runs nothing.

---

## Part 4 · `RacyGet() *Config` — the anti-pattern, on display

The naive double-checked idiom, written deliberately with **no synchronization**:

```
if cfg == nil { cfg = load() }   ← two goroutines both see nil; also an
return cfg                         unsynchronized read/write pair = data race
```

**Pinned contract:** correct under sequential use (non-nil, stable pointer — the graded tests only ever call it from one goroutine). The concurrent hammer lives behind the same opt-in gate as #79:

```
RACE_DEMO=1 go test -race -run RacyGetRaceDemonstration ./80-sync-once/
```

Run it once and watch `-race` flag the exact line — this is the idiom `sync.Once` exists to replace.

---

## Rules

- Raw `sync.Once` for `Get` and `LazyConfig` — note the `sync.OnceValue` form in a comment, don't use it for the drill.
- No mutexes doing Once's job, no `time.Sleep` in your implementations.
- `go vet ./80-sync-once/` clean (copylocks) and `go test -race ./80-sync-once/` clean and green.

## Edge cases the tests hit

- A second `Do` with a different function on the same Once (doesn't run).
- Sequential `Get` calls after the concurrent burst (counter still exactly 1).
- A slow loader with concurrent callers — nobody sees nil (Do blocks).
- A panicking loader: panic propagates once, then permanent nil, zero retries.
- The naive nil-check version under `-race` (opt-in demo).

## What the tests cover

**`Get`/`InitCalls`:** 10-goroutine burst → identical non-nil pointer everywhere and `InitCalls() == 1`; sequential calls afterwards → same pointer, still 1.

**`LazyConfig`:** 20 repeated fresh-instance runs × 10 goroutines → loader count 1 and pointer identity each time; the slow-loader blocking test; the panic sequence (first Get panics, second returns nil calmly, loader ran once).

**`SameOnceTwice`:** first ran, second didn't.

**`RacyGet`:** sequential sanity; the `RACE_DEMO=1`-gated concurrent hammer for the race report.

**Done when:** `-race` is clean, the counter is exactly 1, and you can name the real use cases (DB pools, config, regex compilation) plus the `OnceValue` modern form.
