package util

import (
	"context"

	"github.com/go-logr/logr"
)

func L(ctx context.Context) logr.Logger {
	return logr.FromContextOrDiscard(ctx)
}
