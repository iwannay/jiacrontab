package jiacrontabd

import (
	"io"
	"jiacrontab/pkg/rpc"
)

func call(stack []*pipeCmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}

	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}

		defer func() {
			pipes[0].Close()
			if err == nil {
				err = call(stack[1:], pipes[1:])
			}
			if err != nil {
				// fixed zombie process
				stack[1].Wait()
			}
		}()
	}
	return stack[0].Wait()
}

func rpcCall(serviceMethod string, args, reply interface{}) error {
	return rpc.Call(cfg.AdminAddr, serviceMethod, args, reply)
}
