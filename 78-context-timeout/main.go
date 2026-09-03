package main

import (
	"context"
	"time"
)

func main() {}
func SlowOp(ctx context.Context, d time.Duration) (string, error) {
	select {
	case <-time.After(d):
		return "done", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func SumToN(ctx context.Context, n int64) (int64, error) {
	if n <= 0 {
		return 0, nil
	}
	sum := int64(0)
	for i := int64(1); i <= n; i++ {
		select {
		case <-ctx.Done():
			return int64(sum), ctx.Err()
		default:
		}
		sum += i
	}
	return sum, nil
}

func CallWithTimeout(d time.Duration, op func(ctx context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return op(ctx)
}
