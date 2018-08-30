package kproc

import (
	"context"
	"os/exec"
)

type KCmd struct {
	ctx context.Context
	*exec.Cmd
}
