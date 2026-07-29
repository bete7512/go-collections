# 25 · Invert a map

**Goal:** Swap keys and values. The real lesson is what happens when two keys share a value — and that Go's map iteration order makes "last one wins" nondeterministic.

**Signature (pick one, document why):**
```go
func Invert(m map[string]string) map[string]string              // collisions: unspecified winner
func Invert(m map[string]string) (map[string]string, error)     // collisions: error out
```

**Requirements:**
- Result maps each original value to its original key.
- Confront collisions head-on. If you take the error route, return an error naming the duplicated value. If you take the "unspecified" route, say so in a doc comment — and never write a test asserting which key won.
- Input must not be modified; result is a fresh map. Nil input → empty result, no panic.

**Examples:** `{"a":"1","b":"2"}` → `{"1":"a","2":"b"}`; `{}` → `{}`; `{"a":"x","b":"x"}` → error, or a one-entry map with an unspecified winner.

**Edge cases:** duplicate values (the important one); empty map; nil map; empty-string keys or values.

**Test plan:** clean-invert equality test; collision test matching your documented policy — if "unspecified", assert only `len(result) == 1` and that the winner is one of the two valid keys, never a specific one.

**Done when:** running the collision test 50 times in a loop never flakes.

