package kproc

import (
	"context"
	"fmt"
	"os/exec"
)

func CommandContext(ctx context.Context, name string, arg ...string) *KCmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	return &KCmd{
		Cmd: cmd,
		ctx: ctx,
	}
}

func (k *KCmd) KillAll() {
	if k.Process == nil {
		return
	}
	c := exec.Command("taskkill", "/t", "/f", "/pid", fmt.Sprint(k.Process.Pid))
	c.Stdout = k.Cmd.Stdout
	c.Stderr = k.Cmd.Stderr
}

func (k *KCmd) Wait() error {
	defer k.KillAll()
	go func() {
		select {
		case <-k.ctx.Done():
			k.KillAll()
		}
	}()
	return k.Cmd.Wait()
}
