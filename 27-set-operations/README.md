# 27 · Set operations

**Goal:** Union, intersection, difference on your #26 Set — the operations that make a set type worth having.

**Signature:**
```go
func (s *Set) Union(other *Set) *Set      // everything in either
func (s *Set) Intersect(other *Set) *Set  // only what's in both
func (s *Set) Diff(other *Set) *Set       // in s, not in other  (s − other)
```

**Requirements:**
- Each returns a **new** Set; neither receiver nor argument may be modified. This is the property most implementations get wrong.
- `Diff` direction must be unambiguous: `A.Diff(B)` is A−B. Say it in the doc comment.
- Efficiency note for `Intersect`: iterate the smaller set and probe the larger.
- Handle nil/empty arguments gracefully.

**Examples:** A={1,2,3}, B={2,3,4} → `A.Union(B)`={1,2,3,4}; `A.Intersect(B)`={2,3}; `A.Diff(B)`={1}; `B.Diff(A)`={4}.

**Edge cases:** disjoint sets (intersect empty); identical sets (diff empty); empty receiver; empty argument; self-operation (`A.Union(A)`).

**Test plan:** one test per operation comparing resulting elements (sort them first — sets have no order); **after every operation, assert A and B still have their original contents**; both diff directions; the self-operation cases.

**Done when:** the immutability assertions pass for all three operations.

