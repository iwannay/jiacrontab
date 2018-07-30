package rpc

import (
	"errors"
	"jiacrontab/libs/proto"
	"log"
	"net"
	"net/rpc"
	"time"
)

const (
	diaTimeout   = 5 * time.Second
	callTimeout  = 1 * time.Minute
	pingDuration = 3 * time.Second
)

var (
	ErrRpc        = errors.New("rpc is not available")
	ErrRpcTimeout = errors.New("rpc call timeout")
)

type ClientOptions struct {
	Network string
	Addr    string
}

type Client struct {
	*rpc.Client
	options ClientOptions
	quit    chan struct{}
	err     error
}

func Dial(options ClientOptions) (c *Client) {
	c = &Client{}
	c.options = options
	c.dial()
	return c
}

func (c *Client) dial() (err error) {
	conn, err := net.DialTimeout(c.options.Network, c.options.Addr, diaTimeout)
	if err != nil {
		return err
	}
	c.Client = rpc.NewClient(conn)
	return nil
}

func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if serviceMethod != PingService {
		log.Println("rpc call", c.options.Addr, serviceMethod)
	}

	if c.Client == nil {
		return ErrRpc
	}
	select {
	case call := <-c.Client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done:
		return call.Error
	case <-time.After(callTimeout):
		return ErrRpcTimeout
	}
}

func (c *Client) Error() error {
	return c.err
}

func (c *Client) Close() {
	c.quit <- struct{}{}
}

func (c *Client) Ping(serviceMethod string) {
	var (
		err error
	)
	for {
		select {
		case <-c.quit:
			goto closed
		default:
		}
		if c.Client != nil && c.err == nil {
			if err = c.Call(serviceMethod, &proto.EmptyArgs{}, &proto.EmptyReply{}); err != nil {
				c.err = err
				if err != rpc.ErrShutdown {
					c.Client.Close()
				}
				log.Printf("client.Call(%s, args, reply) error (%v) \n", serviceMethod, err)
			}
		} else {
			if err = c.dial(); err == nil {
				c.err = nil
				log.Println("client reconnet ", c.options.Addr)
			}
		}
		time.Sleep(pingDuration)
	}
closed:
	if c.Client != nil {
		c.Client.Close()
	}
}
