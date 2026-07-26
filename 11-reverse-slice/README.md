# 11 · Reverse a slice in place

- **Build:** `ReverseInPlace(s []int)` — no return value; mutates the argument.
- **Expected:** `[1,2,3,4]` → `[4,3,2,1]`; `[1,2,3]` → `[3,2,1]`.
- **Edge cases:** empty; single element; even vs odd length.
- **Test:** assert the *caller's* slice changed (that's the "in place" proof).
- **Done when:** two-index swap loop, zero allocations (`go test -bench` with `-benchmem` if curious).
