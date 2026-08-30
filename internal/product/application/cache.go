package application

import (
	"context"
	"errors"
)

var ErrCacheMiss = errors.New("cache miss")

type ProductCache interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte) error
	Delete(context.Context, string) error
}
