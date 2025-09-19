package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/admpub/gosshserver"
	"github.com/magiconair/properties"
)

var conf string = `server.properties`
var help bool //= true

func init() {
	flag.StringVar(&conf, `conf`, conf, `-conf server.properties`)
	flag.BoolVar(&help, `help`, help, ``)
	flag.Parse()
}

func main() {
	if help {
		flag.Usage()
		return
	}
	if len(conf) == 0 {
		conf = `server.properties`
	}
	if _, err := os.Stat(conf); err != nil && os.IsNotExist(err) {
		b := bytes.NewBuffer(nil)
		p := properties.NewProperties()
		p.SetValue(`server-ip`, `127.0.0.1`)
		p.SetValue(`server-port`, `10022`)
		p.SetValue(`term-user`, `root`)
		p.SetValue(`term-password`, ``)
		p.SetValue(`term-key-path`, `ssh.key`)
		p.SetValue(`trusted-user-ca-keys[0]`, ``)
		p.SetComment(`trusted-user-ca-keys[0]`, `# trusted-user-ca-keys 支持多个，格式：file:/path/to/ca.pub 或 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAr... user@host`)
		p.Write(b, properties.UTF8)
		os.WriteFile(conf, b.Bytes(), 0664)
	}
	cfg := gosshserver.Config{}
	// 解析 server.properties
	err := properties.MustLoadFile(conf, properties.UTF8).Decode(&cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = cfg.SetDefaults()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	gosshserver.Serve(cfg, nil)
}
