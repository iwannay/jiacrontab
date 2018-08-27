// +build !windows
package kproc

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func CommandContext(ctx context.Context, name string, arg ...string) *KCmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true
	return &KCmd{
		ctx: ctx,
		Cmd: cmd,
	}
}

func (k *KCmd) KillAll() {
	if k.Process == nil {
		return
	}
	group, err := os.FindProcess(-k.Process.Pid)
	if err == nil {
		err = group.Signal(syscall.SIGKILL)
	}
	if err != nil {
		log.Println("process.Wait error:", err)
	}
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
