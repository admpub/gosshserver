package gosshserver

import (
	"fmt"
	"io"
	"log"

	"github.com/admpub/gopty"
	"github.com/gliderlabs/ssh"
)

func shellHandler(s ssh.Session) {
	ptyReq, winCh, isPty := s.Pty()
	// if !isPty {
	// 	fmt.Fprintln(s, "Must be PTY")
	// 	s.Exit(1)
	// 	return
	// }
	_ = isPty

	pty, err := gopty.New(40, 30)
	if err != nil {
		fmt.Fprintln(s, "Can not start shell")
		log.Println(err)
		s.Exit(1)
		return
	}

	// set environment variables
	if len(ptyReq.Term) > 0 {
		pty.SetENV([]string{fmt.Sprintf("TERM=%s", ptyReq.Term)})
	}

	defer func() { _ = pty.Close() }()
	err = gopty.Start(pty)

	go func() {
		for win := range winCh {
			pty.SetSize(int(win.Width), int(win.Height))
		}
	}()

	go func() {
		io.Copy(s, pty)
	}()

	go func() {
		io.Copy(pty, s)
		s.Close()
	}()

	// == wait ==

	if _, err := pty.Wait(); err != nil {
		s.Exit(1)
	} else {
		s.Exit(0)
	}
}
