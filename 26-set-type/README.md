# 26 · Set type

**Goal:** Build the set Go's stdlib doesn't give you, on top of `map[T]struct{}`. Foundation for #27 and #64.

**Signature:**
```go
type Set struct{ m map[string]struct{} }   // or: type Set map[string]struct{}

func NewSet(items ...string) *Set
func (s *Set) Add(v string)
func (s *Set) Has(v string) bool
func (s *Set) Remove(v string)
func (s *Set) Len() int
```

**Requirements:**
- `struct{}{}` as the value type — it occupies zero bytes, unlike `bool`. Be able to say that out loud.
- `Add` of an existing element is a no-op (length unchanged); `Remove` of an absent element is a no-op, never a panic (Go's `delete` is already safe on missing keys).
- Decide whether the zero value `var s Set` is usable or `NewSet` is mandatory. If mandatory, make `Add` on a nil map fail loudly rather than panic cryptically — or lazily initialize. Document either way.
- Variadic constructor so `NewSet("a","b")` works.

**Examples:** `NewSet("a","a","b")` → `Len()==2`; `Has("a")` true, `Has("z")` false; after `Remove("a")` → `Len()==1`, `Has("a")` false.

**Edge cases:** double Add; Remove absent; Has on empty set; Remove down to empty; zero-value usage.

**Test plan:** one lifecycle test (add → has → remove → has) plus dedicated double-add and remove-absent tests; a `Len` assertion after every mutation.

**Done when:** all pass and you can state the `struct{}` vs `bool` memory argument.

