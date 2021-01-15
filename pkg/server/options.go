package server

import (
	"golang.org/x/crypto/ssh"
)

type Option func(*Options)

type Options struct {
	Addr            string
	Port            int
	PrometheusAddr  string
	PrometheusPort  int
	SSHServerConfig *ssh.ServerConfig
}

func (o *Options) Apply(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

func WithPrometheusAddr(addr string) Option {
	return func(o *Options) {
		o.PrometheusAddr = addr
	}
}

func WithPrometheusPort(port int) Option {
	return func(o *Options) {
		o.PrometheusPort = port
	}
}

func WithSSHServerConfig(cfg *ssh.ServerConfig) Option {
	return func(o *Options) {
		o.SSHServerConfig = cfg
	}
}
