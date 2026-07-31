# 29 · Deduplicate a slice of structs by one field

## Problem

You have a slice of structs and some of them are duplicates — not identical structs, but structs sharing the same value in one identifying field. Return a new slice with the duplicates removed, keeping only the **first** occurrence of each ID. This is the everyday shape of cleaning up rows joined from multiple sources, merged API responses, or replayed events keyed by entity ID.

## Contract (what the tests enforce)

```go
type User struct {
    ID   int
    Name string
}

func DedupByID(users []User) []User
```

- **First occurrence wins.** When two elements share an ID, the one at the lower index survives — including its *other* fields. `[{1,"a"},{1,"b"}]` keeps `{1,"a"}`, never `{1,"b"}`. Duplicates are judged **only by ID**; the Name field plays no part in equality.
- **Order preserved.** Survivors appear in the same relative order as in the input.
- **Input untouched.** The input slice must be exactly the same after the call. Return fresh memory (or at minimum, never write into the input's backing array).
- **Empty and nil inputs** return an empty result, no panic. (The tests compare with `slices.Equal`, which treats nil and empty as equal — either is accepted.)
- **Zero and negative IDs are legitimate.** `0` is a valid ID and must dedup like any other value — an implementation that abuses `0` as a sentinel will fail.
- One pass, O(n): a `map[int]struct{}` of seen IDs plus an output slice. (The tests can't measure this directly, but the 10,000-element case will feel it.)

## Worked examples

**Example 1 — basic dedup, first wins:**

```
input:  [{1 "alice"} {2 "bob"} {1 "carol"}]
output: [{1 "alice"} {2 "bob"}]
```

ID 1 appears twice; index 0 wins, so `"alice"` survives and `"carol"` is dropped — even though the structs differ.

**Example 2 — non-adjacent duplicates, order kept:**

```
input:  [{3 "x"} {1 "y"} {3 "z"} {2 "w"} {1 "q"}]
output: [{3 "x"} {1 "y"} {2 "w"}]
```

Duplicates don't have to be neighbors. Survivors keep their original relative order: 3, then 1, then 2.

**Example 3 — nothing to do:**

```
input:  [{1 "a"} {2 "b"} {3 "c"}]
output: [{1 "a"} {2 "b"} {3 "c"}]
```

No shared IDs → output equals input (as a copy, not the same backing array).

## Edge cases the tests cover

- Empty slice and nil slice → empty result.
- Single element → returned as-is.
- **All** elements share one ID → single-element result, first one.
- Zero-value ID (`0`) deduping correctly.
- Negative IDs.
- Duplicate IDs where every Name differs — proves Name is ignored for equality and first-wins is honored.
- Three or more occurrences of the same ID — still only the first survives.
- Input-unmodified check: after the call, the original slice must be byte-for-byte what it was.
- A 10,000-element input with heavy duplication — a correctness stress, and quadratic solutions will be noticeably slow.
