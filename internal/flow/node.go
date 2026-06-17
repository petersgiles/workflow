package flow

import (
	"errors"
)

// ErrAborted is returned by orchestrator when flow stops.
var ErrAborted = errors.New("flow aborted")
