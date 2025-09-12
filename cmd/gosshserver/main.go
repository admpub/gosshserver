package main

import (
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
		os.WriteFile(conf, nil, 0664)
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
