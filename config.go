package gosshserver

import (
	"os"
	"strings"

	gossh "golang.org/x/crypto/ssh"
)

type Config struct {
	ServerIP                string            `properties:"server-ip,default=127.0.0.1"`
	ServerPort              string            `properties:"server-port,default=10022"`
	User                    string            `properties:"term-user,default=root"`
	Password                string            `properties:"term-password,default="`
	KeyPath                 string            `properties:"term-key-path,default=ssh.key"`
	LocalPortForwarding     bool              `properties:"local-port-forwarding,default=false"`
	ReversePortForwarding   bool              `properties:"reverse-port-forwarding,default=false"`
	TrustedUserCAKeys       []string          `properties:"trusted-user-ca-keys,default="`
	TrustedUserCAKeysParsed []gossh.PublicKey `properties:"-"`
	initialized             bool              `properties:"-"`
}

func (c *Config) SetDefaults() error {
	if c.initialized {
		return nil
	}
	c.initialized = true
	var err error
	if len(c.TrustedUserCAKeys) > 0 {
		for _, pk := range c.TrustedUserCAKeys {
			pk = strings.TrimSpace(pk)
			if len(pk) == 0 {
				continue
			}
			// 支持 file:/path/to/ca.pub 或 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAr... user@host
			var data []byte
			if cut, found := strings.CutPrefix(pk, `file:`); found {
				data, err = os.ReadFile(cut)
				if err != nil {
					return err
				}
			} else {
				data = []byte(pk)
			}
			var publicKey gossh.PublicKey
			publicKey, err = parsePublicKeyPEM(data)
			if err != nil {
				return err
			}
			c.TrustedUserCAKeysParsed = append(c.TrustedUserCAKeysParsed, publicKey)
		}
	}
	if len(c.KeyPath) == 0 {
		c.KeyPath = `ssh.key`
	}
	return err
}
