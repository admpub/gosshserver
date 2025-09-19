package gosshserver

import (
	"fmt"
	"net"

	"github.com/admpub/log"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func Serve(cfg Config, privateKeyPEM []byte) error {
	err := cfg.SetDefaults()
	if err != nil {
		return err
	}
	if privateKeyPEM == nil {
		if err := initPrivatePEM(cfg.KeyPath); err != nil {
			return err
		}

		privateKeyPEM = privatePEM
	}
	// 创建 ssh 服务器密钥
	privateKeySigner, err := gossh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return fmt.Errorf("不能解析私钥: %w", err)
	}

	// 在指定端口开启服务
	address := net.JoinHostPort(cfg.ServerIP, cfg.ServerPort)

	s := &ssh.Server{
		Addr:             address,
		Handler:          shellHandler,
		PublicKeyHandler: publicKeyHandler(cfg),
		PasswordHandler:  passwordHandler(cfg),
		Version:          `0.0.1`,
		Banner:           `欢迎使用 SSH 服务` + "\n",
		BannerHandler:    bannerHandler,
		LocalPortForwardingCallback: func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		},
	}
	s.AddHostKey(privateKeySigner)

	log.Okay("Server Address:", address)
	if err = s.ListenAndServe(); err != nil {
		log.Errorf("不能启动服务器: %v", err)
	}
	return err
}
