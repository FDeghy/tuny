package tcp

import (
	"time"

	"github.com/rs/zerolog"
)

var (
	logger zerolog.Logger
)

const (
	MaxWriteBufferSize = 4 * 1024 * 1024
	MaxIdleConnection  = 2 * time.Minute
)
