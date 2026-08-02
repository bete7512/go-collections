# 35 · Sum transactions per user

## Problem

Given a list of transactions, produce each user's total. This is the most common aggregation shape in backend work — GROUP BY + SUM done in application code — and the last drill of the maps tier. It also carries a lesson about money and floating point that the tests make you feel.

## Contract (what the tests enforce)

```go
type Tx struct {
    UserID string
    Amount float64
}

func Totals(txs []Tx) map[string]float64
```

- **One entry per distinct UserID**, holding the sum of all that user's amounts, in input order of accumulation (order doesn't affect the sum; one pass is the expected shape).
- **Negative amounts are refunds** and subtract normally.
- **A user whose transactions net to exactly zero still appears in the result** with total 0. Having transacted is information; an implementation that drops zero-sum users fails.
- **Users appear only if they have at least one transaction** — no phantom keys.
- Empty or nil input → empty result (the tests accept nil or empty).
- The empty-string UserID `""` is a legitimate key and aggregates like any other.
- The input slice is not modified.
- **Float comparisons in the tests use a tolerance** (`|got-want| < 1e-9`), not `==` — because `0.1+0.2 != 0.3` in binary floating point. Your implementation just sums `float64`s; the tolerance is the test's job. But the takeaway is required reading: production money code uses integer minor units (cents) or a decimal type, never `float64`. Write that in a comment.

## Worked examples

**Example 1 — basic aggregation:**

```
input: [{u1 10} {u2 5} {u1 2.5}]
→ {"u1": 12.5, "u2": 5}
```

u1 has two transactions summing to 12.5; u2 has one.

**Example 2 — refunds, and a net-zero user:**

```
input: [{u1 100} {u1 -30} {u2 20} {u2 -20}]
→ {"u1": 70, "u2": 0}
```

u2 bought and fully refunded — the total is 0, but u2 is present. `Get`ting u2 from the map must find the key.

**Example 3 — why the tolerance exists:**

```
input: ten transactions of {u1 0.1}
→ {"u1": ~1.0}   // actually 0.9999999999999999 in float64
```

Summing 0.1 ten times does not produce exactly 1.0 — 0.1 has no exact binary representation. The test asserts the total is within 1e-9 of 1.0, and that's the moment you internalize why money isn't a float.

## Edge cases the tests cover

- Empty and nil input.
- Single transaction; single user with many transactions.
- Refunds (negative amounts), including a mixed positive/negative stream.
- A user netting to exactly zero — present in the result.
- Zero-amount transactions (user must appear).
- The empty-string UserID.
- Fractional amounts accumulating float error (tolerance-checked).
- Many users interleaved — accumulation must be keyed, not positional.
- A 10,000-transaction input across 100 users with known per-user totals.
- Input-unmodified snapshot.
