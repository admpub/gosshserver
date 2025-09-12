package gosshserver

import "github.com/gliderlabs/ssh"

func passwordHandler(cfg Config) func(ctx ssh.Context, password string) bool {
	return func(ctx ssh.Context, password string) bool {
		if len(cfg.Password) == 0 {
			return false
		}
		return ctx.User() == cfg.User && password == cfg.Password
	}
}
