package server

import (
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
)

// ListenAndServe announces addr on the local network, accepts incoming connections,
// transfers commands from connection to handler and command execution results
// from handler to connection
func ListenAndServe(addr string, handler Handler, cfg *Config) error {
	if handler == nil {
		return errors.New("invalid server handler")
	}

	// default address
	if addr == "" {
		addr = ":7777"
	}

	// setup log
	var logger *log.Logger
	if cfg == nil || cfg.Logger == nil {
		logger = log.New(ioutil.Discard, "", 0)
	} else {
		logger = cfg.Logger
	}

	// start listening server socket
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger.Println("listening ", listener.Addr())

	stopped := false
	if cfg != nil && cfg.Stop != nil {
		// running stop monitor
		go func() {
			<-cfg.Stop
			stopped = true
			logger.Println("received stop signal...")
			listener.Close()
		}()
	}

	for {
		netconn, err := listener.Accept()
		if err != nil {
			if stopped {
				return nil
			}
			return err
		}

		conn := &connection{textproto.NewConn(netconn),
			netconn.RemoteAddr(),
			logger,
		}
		conn.log("connected")

		go conn.serve(handler)
	}
}
