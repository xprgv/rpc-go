package rpc

import (
	"context"
	"time"
)

func getTimeoutFromContext(ctx context.Context) time.Duration {
	duration := time.Duration(0)
	if deadline, ok := ctx.Deadline(); ok {
		duration = time.Until(deadline)
	}
	return duration
}
