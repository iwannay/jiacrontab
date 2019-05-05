// +build !windows

package kproc

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"syscall"

	"github.com/iwannay/log"
)

func CommandContext(ctx context.Context, name string, arg ...string) *KCmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true
	return &KCmd{
		ctx:                ctx,
		Cmd:                cmd,
		isKillChildProcess: true,
		done:               make(chan struct{}),
	}
}

// SetUser 设置执行用户要保证root权限
func (k *KCmd) SetUser(username string) {
	if username == "" {
		return
	}
	u, err := user.Lookup(username)
	if err == nil {

		log.Infof("KCmd set uid=%s,gid=%s", u.Uid, u.Gid)
		k.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUSER,
		}
		// 以下代码不能切换用户
		// uid, _ := strconv.Atoi(u.Uid)
		// gid, _ := strconv.Atoi(u.Gid)
		// k.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
}

func (k *KCmd) KillAll() {

	select {
	case k.done <- struct{}{}:
	default:
	}

	if k.Process == nil {
		return
	}

	if k.isKillChildProcess == false {
		return
	}

	group, err := os.FindProcess(-k.Process.Pid)
	if err == nil {
		group.Signal(syscall.SIGKILL)
	}
}

func (k *KCmd) Wait() error {
	defer k.KillAll()
	go func() {
		select {
		case <-k.ctx.Done():
			k.KillAll()
		case <-k.done:
		}
	}()
	return k.Cmd.Wait()
}
