package main

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

func withGuard(t *testing.T, d time.Duration, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s did not finish within %v — the operation must observe ctx and return early", what, d)
	}
}

func canceledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func expiredCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), -time.Millisecond)
	_ = cancel // deadline already passed; Done is closed on creation
	return ctx
}

// ---------- SlowOp ----------

func TestSlowOpCompletesUnderGenerousTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	var v string
	var err error
	withGuard(t, 3*time.Second, "SlowOp(50ms work, 2s timeout)", func() {
		v, err = SlowOp(ctx, 50*time.Millisecond)
	})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if v != "done" {
		t.Errorf("value = %q, want %q", v, "done")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("took %v — 50ms of work must not take over a second", elapsed)
	}
}

func TestSlowOpNoDeadlineContext(t *testing.T) {
	var v string
	var err error
	withGuard(t, 3*time.Second, "SlowOp with context.Background()", func() {
		v, err = SlowOp(context.Background(), 30*time.Millisecond)
	})

	if err != nil || v != "done" {
		t.Errorf("got (%q, %v), want (%q, nil)", v, err, "done")
	}
}

func TestSlowOpTimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	start := time.Now()
	var v string
	var err error
	withGuard(t, 3*time.Second, "SlowOp(2s work, 80ms timeout)", func() {
		v, err = SlowOp(ctx, 2*time.Second)
	})
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
	if errors.Is(err, context.Canceled) {
		t.Errorf("err = %v — a fired deadline is DeadlineExceeded, not Canceled", err)
	}
	if v != "" {
		t.Errorf("value = %q, want empty string on error", v)
	}
	// The proof of early return: the error alone could be produced after
	// waiting out the full 2s of "work".
	if elapsed > time.Second {
		t.Errorf("took %v — an 80ms deadline must abandon 2s work early", elapsed)
	}
}

func TestSlowOpExplicitCancelBeatsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type outcome struct {
		v   string
		err error
	}
	got := make(chan outcome, 1)
	start := time.Now()
	go func() {
		v, err := SlowOp(ctx, 2*time.Second)
		got <- outcome{v, err}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel() // explicit cancel long before the 5s deadline

	select {
	case out := <-got:
		if !errors.Is(out.err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled for an explicit cancel", out.err)
		}
		if errors.Is(out.err, context.DeadlineExceeded) {
			t.Errorf("err = %v — cancel() before the deadline is Canceled, not DeadlineExceeded", out.err)
		}
		if out.v != "" {
			t.Errorf("value = %q, want empty string on error", out.v)
		}
		if elapsed := time.Since(start); elapsed > time.Second {
			t.Errorf("took %v — cancel at 50ms must not wait out the 2s work", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("SlowOp did not return after cancel()")
	}
}

func TestSlowOpPreExpiredContext(t *testing.T) {
	start := time.Now()
	var err error
	withGuard(t, 3*time.Second, "SlowOp with an expired context", func() {
		_, err = SlowOp(expiredCtx(), 2*time.Second)
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("took %v — an already-expired context must return immediately", elapsed)
	}
}

func TestSlowOpPreCanceledContext(t *testing.T) {
	var err error
	withGuard(t, 3*time.Second, "SlowOp with a pre-canceled context", func() {
		_, err = SlowOp(canceledCtx(), 2*time.Second)
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestSlowOpTimeoutIsDeterministic(t *testing.T) {
	for run := 0; run < 5; run++ {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)

		var err error
		withGuard(t, 3*time.Second, "SlowOp determinism run", func() {
			_, err = SlowOp(ctx, 2*time.Second)
		})
		cancel()

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("run %d: err = %v, want context.DeadlineExceeded", run, err)
		}
	}
}

// ---------- SumToN ----------

func TestSumToNSmall(t *testing.T) {
	var sum int64
	var err error
	withGuard(t, 3*time.Second, "SumToN(10)", func() {
		sum, err = SumToN(context.Background(), 10)
	})

	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if sum != 55 {
		t.Errorf("SumToN(10) = %d, want 55", sum)
	}
}

func TestSumToNZeroAndNegative(t *testing.T) {
	for _, n := range []int64{0, -5} {
		var sum int64
		var err error
		withGuard(t, 3*time.Second, "SumToN(n<=0)", func() {
			sum, err = SumToN(context.Background(), n)
		})
		if sum != 0 || err != nil {
			t.Errorf("SumToN(%d) = (%d, %v), want (0, nil)", n, sum, err)
		}
	}
}

func TestSumToNAbandonsOnTimeout(t *testing.T) {
	// A pure CPU loop never blocks, so ctx can only stop it if the loop
	// itself checks ctx.Done(). Without that check this call runs for
	// centuries and the guard fails the test.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	start := time.Now()
	var err error
	withGuard(t, 5*time.Second, "SumToN(math.MaxInt64, 60ms timeout)", func() {
		_, err = SumToN(ctx, math.MaxInt64)
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("took %v — the loop must notice the deadline within, not after, the work", elapsed)
	}
}

func TestSumToNPreCanceled(t *testing.T) {
	var err error
	withGuard(t, 3*time.Second, "SumToN with a pre-canceled context", func() {
		_, err = SumToN(canceledCtx(), math.MaxInt64)
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// ---------- CallWithTimeout ----------

func TestCallWithTimeoutPassesResultsVerbatim(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var v string
		var err error
		withGuard(t, 3*time.Second, "CallWithTimeout with a fast op", func() {
			v, err = CallWithTimeout(2*time.Second, func(ctx context.Context) (string, error) {
				return "ok", nil
			})
		})
		if v != "ok" || err != nil {
			t.Errorf("got (%q, %v), want (%q, nil)", v, err, "ok")
		}
	})

	t.Run("op's own error", func(t *testing.T) {
		boom := errors.New("boom")
		var err error
		withGuard(t, 3*time.Second, "CallWithTimeout with a failing op", func() {
			_, err = CallWithTimeout(2*time.Second, func(ctx context.Context) (string, error) {
				return "", boom
			})
		})
		if !errors.Is(err, boom) {
			t.Errorf("err = %v, want the op's own error passed through verbatim", err)
		}
	})
}

func TestCallWithTimeoutHandsOpALiveDeadlineContext(t *testing.T) {
	before := time.Now()
	withGuard(t, 3*time.Second, "CallWithTimeout context inspection", func() {
		CallWithTimeout(2*time.Second, func(ctx context.Context) (string, error) {
			if err := ctx.Err(); err != nil {
				t.Errorf("ctx.Err() = %v on entry, want nil — the context must still be live", err)
			}
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Errorf("ctx.Deadline() reports no deadline — op must receive a WithTimeout context, not a bare Background")
			} else if deadline.Before(before) {
				t.Errorf("deadline %v is already in the past", deadline)
			}
			return "ok", nil
		})
	})
}

func TestCallWithTimeoutExpires(t *testing.T) {
	slowOp := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(2 * time.Second):
			return "late", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	start := time.Now()
	var v string
	var err error
	withGuard(t, 3*time.Second, "CallWithTimeout(60ms) with a 2s op", func() {
		v, err = CallWithTimeout(60*time.Millisecond, slowOp)
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
	if v != "" {
		t.Errorf("value = %q, want empty string", v)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("took %v — a 60ms timeout must abandon a 2s op early", elapsed)
	}
}

func TestCallWithTimeoutZeroDuration(t *testing.T) {
	ctxAware := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(2 * time.Second):
			return "late", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	start := time.Now()
	var err error
	withGuard(t, 3*time.Second, "CallWithTimeout(0)", func() {
		_, err = CallWithTimeout(0, ctxAware)
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded — d<=0 expires the context immediately", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("took %v — a zero timeout must return immediately", elapsed)
	}
}
