package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/peertechde/zodiac/pkg/logging"
)

var log = logging.Logger.WithField("subsys", "ssh")

var (
	ErrInitiliazed    = fmt.Errorf("already initialized")
	ErrNotInitiliazed = fmt.Errorf("not  initialized")
)

type trackingConn struct {
	net.Conn
}

func (tc *trackingConn) Read(b []byte) (int, error) {
	n, err := tc.Conn.Read(b)
	sshReadCalls.Inc()
	sshReadBytes.Add(float64(n))

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			sshReadTimeout.Inc()
		} else {
			sshReadErrors.Inc()
		}
	}
	return n, err
}

func (tc *trackingConn) Write(b []byte) (int, error) {
	n, err := tc.Conn.Write(b)
	sshWriteCalls.Inc()
	sshWriteBytes.Add(float64(n))

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			sshReadTimeout.Inc()
		} else {
			sshWriteErrors.Inc()
		}
	}
	return n, err
}

func New() *Server {
	return &Server{
		initialized: make(chan struct{}),
	}
}

type Server struct {
	initializeOnce sync.Once
	initialized    chan struct{}

	options Options
	wg      sync.WaitGroup
}

func (s *Server) Initialize(options ...Option) error {
	err := ErrInitiliazed
	s.initializeOnce.Do(func() {
		close(s.initialized)
		err = s.initialize(options...)
	})
	return err
}

func (s *Server) initialize(options ...Option) error {
	s.options.Apply(options...)

	return nil
}

func (s *Server) isInitialized() bool {
	select {
	case <-s.initialized:
		return true
	default:
	}
	return false
}

func (s *Server) Serve(ctx context.Context) error {
	if !s.isInitialized() {
		return ErrNotInitiliazed
	}

	addr := net.JoinHostPort(s.options.Addr, strconv.Itoa(s.options.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	promAddr := net.JoinHostPort(s.options.PrometheusAddr, strconv.Itoa(s.options.PrometheusPort))
	promLn, err := net.Listen("tcp", promAddr)
	if err != nil {
		return err
	}
	promServer := &http.Server{
		Handler: promhttp.Handler(),
	}

	var group run.Group
	{
		group.Add(
			func() error {
				for {
					conn, err := ln.Accept()
					if err != nil {
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							continue
						}
						return err
					}
					s.wg.Add(1)
					go s.handleConn(conn)
				}
				s.wg.Wait()

				return nil
			},
			func(err error) {

			},
		)
	}
	{
		group.Add(
			func() error {
				if err := promServer.Serve(promLn); err != nil {
					return err
				}
				return nil
			},
			func(err error) {

			},
		)
	}
	if err := group.Run(); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	sshConnsAttemptsTotal.Inc()

	tConn := &trackingConn{Conn: conn}
	sconn, chans, reqs, err := ssh.NewServerConn(tConn, s.options.SSHServerConfig)
	if err != nil {
		sshConnsAttemptsErrors.Inc()
		s.wg.Done()

		return
	}
	scopedLog := log.WithFields(logrus.Fields{
		"remote":         sconn.RemoteAddr(),
		"user":           sconn.User(),
		"client-version": string(sconn.ClientVersion()),
	})
	scopedLog.Info("handling connection")

	defer func() {
		s.wg.Done()
		sshConnsOpen.Dec()

		scopedLog.Info("closed connection")
	}()
	sshConnsOpen.Inc()
	sshConnsTotal.Inc()

	// TODO: implement some kind of context cancellation...
	for {
		select {
		case req := <-reqs:
			if req == nil {
				return
			}
			// discard requests for now
			if req.WantReply {
				req.Reply(false, nil)
			}
		case ch := <-chans:
			if ch == nil {
				return
			}
			go s.handleChannel(sconn, ch)
		}
	}
}

func (s *Server) handleChannel(sconn *ssh.ServerConn, channel ssh.NewChannel) {
	switch channel.ChannelType() {
	case "direct-tcpip":
		s.handleProxyJump(sconn, channel)
	case "session":
		// TODO
	default:
		channel.Reject(ssh.UnknownChannelType, "connection type not supported")
	}
}

// direct-tcpip message data as specified in RFC4254
type openDirectMsg struct {
	RemoteAddr string
	RemotePort uint32
	SourceAddr string
	SourcePort uint32
}

func (s *Server) handleProxyJump(sconn *ssh.ServerConn, channel ssh.NewChannel) {
	scopedLog := log.WithFields(logrus.Fields{
		"remote":         sconn.RemoteAddr(),
		"user":           sconn.User(),
		"client-version": string(sconn.ClientVersion()),
	})
	scopedLog.Info("handling proxy-jump")

	var msg openDirectMsg
	if err := ssh.Unmarshal(channel.ExtraData(), &msg); err != nil {
		scopedLog.Error("failed to parse direct-tcpip request (%v)", err)
		channel.Reject(ssh.UnknownChannelType, "failed to parse direct-tcpip request")
		return
	}
	ch, reqs, err := channel.Accept()
	if err != nil {
		scopedLog.Errorf("failed to accept channel (%v)", err)
		channel.Reject(ssh.ConnectionFailed, "unable to accept channel")
		return
	}
	go ssh.DiscardRequests(reqs)

	raddr := net.JoinHostPort(msg.RemoteAddr, strconv.Itoa(int(msg.RemotePort)))
	conn, err := net.Dial("tcp", raddr)
	if err != nil {
		scopedLog.WithField("raddr", raddr).Errorf("failed to establish a connection (%v)", err)
		channel.Reject(ssh.ConnectionFailed, "failed to establish connection to the endpoint")
		return
	}

	var closeOnce sync.Once
	closer := func() {
		ch.Close()
		conn.Close()
	}
	go func() {
		io.Copy(ch, conn)
		closeOnce.Do(closer)
	}()
	io.Copy(conn, ch)
	closeOnce.Do(closer)
}
