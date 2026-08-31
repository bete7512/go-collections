# 75 · Fan-in (merge N channels)

## The idea in one paragraph

**Fan-in** is the mirror of #74: many channels in, one channel out. Where fan-out was free (N receivers on one channel), fan-in requires actual work — you cannot merge channels by wishing. You need **one forwarding goroutine per input channel**, all sending into a shared output, plus something that closes the output once every forwarder is finished.

Unlike the previous challenges, `Merge` here is a **reusable library function**, not a demo. This exact shape ships in production Go services constantly: merging results from several backends, combining event streams, collecting from a pool of producers. Write it once, properly, and you'll copy it for years.

---

## The functions

| function | what it teaches |
|---|---|
| `Merge` | the core pattern: one forwarder per input + a closer goroutine |
| `MergeCollect` | a convenience drain (what callers usually want) |
| `MergeWithLabels` | preserving per-source identity through the merge |

---

## Function 1 · `Merge(chans ...<-chan int) <-chan int`

**What it gets:** any number of input channels (zero, one, or many). Someone else feeds and closes them.

**What it must return:** a single receive-only channel carrying **every value from every input**, which **closes** once all inputs are drained and closed. It must return this channel **immediately** — before any values have been forwarded.

**Step by step:**

1. Create the output channel and a `sync.WaitGroup`.
2. **For each input channel, `wg.Add(1)` and start a forwarder goroutine.** Each forwarder does one thing: `for v := range itsInput { out <- v }`, then `wg.Done()`. It ranges its *own* input only — it doesn't know the others exist.
3. **Start one closer goroutine:** `wg.Wait()`, then `close(out)`. This is the same closer pattern from #73, and it exists for the same reason — the caller must be free to start receiving immediately, so `Wait` cannot happen on the calling goroutine.
4. Return `out` **now**, without waiting for anything.

**What returning `<-chan int` buys you:** the caller physically cannot close your output channel or send into it. The forwarders and the closer own it, and the type system enforces that.

**The two failure modes you're avoiding:**

- **Close too early** (e.g. closing `out` right after starting the forwarders): the forwarders panic with *send on closed channel*.
- **Never close**: the caller's `range` never ends. This is why the closer goroutine must exist and why every test here is timeout-guarded.

**Traced example:**

```
Merge(a, b, c)   where  a: 1 2      b: 10 20 30      c: 100

  forwarder-a ──┐
  forwarder-b ──┼──→  out  ──→ caller
  forwarder-c ──┘

  each forwarder ranges its own channel and sends into out
  a closes  → forwarder-a returns → wg 3→2
  c closes  → forwarder-c returns → wg 2→1
  b closes  → forwarder-b returns → wg 1→0
                                    closer: close(out)
  caller's range ends

  → 6 values total, interleaved unpredictably:
    e.g. [1 10 100 20 2 30]  or  [10 1 20 30 100 2]  — both correct
```

**Order guarantees:** none *between* sources. But **within** one source, order is preserved — that source's forwarder receives and sends sequentially. The tests check exactly that: each input's values must appear in the output in their original relative order.

**Edge cases:** `Merge()` with **no arguments** must return an **already-closed** channel — the WaitGroup is at zero, so the closer fires immediately. A caller ranging it gets nothing and exits, rather than blocking forever. `Merge(single)` behaves like a pass-through. A **nil channel** among the inputs would block its forwarder forever (nil channels never become ready), so filter nils out before starting forwarders — the tests pass one in.

---

## Function 2 · `MergeCollect(chans ...<-chan int) []int`

**What it gets:** the same variadic inputs.

**What it must do:** call `Merge`, range the merged channel until it closes, return everything as a slice. Three lines on top of Function 1.

**What it must return:** all values from all inputs, in arrival order, as an empty non-nil slice when there's nothing.

**Why it exists:** it's the shape callers actually want most of the time, and it proves `Merge` really does close — if it didn't, this function would hang forever, which is exactly what the timeout guard reports.

---

## Function 3 · `MergeWithLabels(sources map[string]<-chan int) []Labeled`

```go
type Labeled struct {
    Source string
    Value  int
}
```

**What it gets:** named channels — `{"api": ch1, "cache": ch2, "db": ch3}`.

**What it must do:**

1. Same structure as `Merge`, but each forwarder knows its source's **name** and sends `Labeled{Source: name, Value: v}` instead of a bare int.
2. Closer goroutine closes the labeled output; collect the whole stream and return it.

**What it must return:** one `Labeled` per input value, with the correct source name attached.

**Why this matters:** a merged stream usually loses provenance — once values are combined you can't tell where they came from. Attaching the label at the forwarder (the only place that still knows) is how real pipelines keep it. It's also what makes per-source ordering *testable*: the tests group by `Source` and verify each group's values are still in their original order.

**A Go detail worth noticing:** you're iterating a map to start the forwarders, and map iteration order is random — but that's harmless here, since starting order has no bearing on correctness. Contrast with #31, where iteration order mattered enormously. Knowing which situation you're in is the skill.

**Traced example:**

```
MergeWithLabels({"api": [1,2], "db": [10,20]})

  → [{api 1} {db 10} {api 2} {db 20}]     ← one possible interleaving
  → [{db 10} {db 20} {api 1} {api 2}]     ← equally valid

  Grouped by source, though:
     api → [1 2]     always in this order
     db  → [10 20]   always in this order
```

---

## Rules that apply to all three

- **Return the merged channel immediately;** never block the caller while forwarding.
- **One goroutine per input, one closer, no exceptions.** No counting values, no polling.
- **Every value from every input appears exactly once** in the output.
- **Per-source order is preserved; cross-source order is unspecified** and never asserted.
- **Zero inputs → closed channel / empty result.** **One input → pass-through.** **Nil channels are skipped.**
- **Nothing leaks:** after the merged channel closes, every forwarder and the closer have returned. One test checks the goroutine count returns to baseline.
- **No `time.Sleep`.** **`go test -race ./75-fan-in/`** must be clean.

---

## What the tests cover

**`Merge` / `MergeCollect`:** three channels of differing lengths; a single channel; zero channels (immediately-closed output); channels that are empty-but-closed; one long channel (1,000 values) alongside short ones; a nil channel mixed in with real ones; and 8 channels × 500 values each, all 4,000 accounted for exactly once.

**Ordering:** per-source relative order preserved, verified by extracting each source's subsequence from the merged output.

**Termination:** the merged channel actually closes (proved by the range returning), including the zero-input case — everything timeout-guarded so a missing close fails in seconds rather than hanging.

**Leaks:** goroutine count before vs. after a full merge, with a settle delay.

**Repetition:** 50 runs of the same merge, each accounting for every value — concurrency bugs pass single runs.

**`MergeWithLabels`:** correct source attached to every value; per-source ordering within groups; an empty map; a map with one source; a source with no values; and 100 values across 5 sources.
