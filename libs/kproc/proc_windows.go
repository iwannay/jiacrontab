package kproc

import (
	"context"
	"os/exec"
)

func CommandContext(ctx context.Context, name string, arg ...string) *KCmd {
	return &KCmd{
		Cmd: cmd,
		ctx: ctx,
	}
}

func (k *KCmd) KillAll() {
	if k.Process == nil {
		return
	}
	c := exec.Command("taskkill", "/t", "/f", "/pid", k.Process.Pid)
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
