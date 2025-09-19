package gosshserver

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

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

	// == welcome message ==

	welcome := []string{
		time.Now().Format(time.RFC3339),
		`Welcome ` + s.User() + ", " + s.RemoteAddr().String(),
	}
	var maxLen int
	for _, line := range welcome {
		maxLen = max(maxLen, len(line)+2)
	}
	headAndFoot := `+` + strings.Repeat(`-`, maxLen) + "+\n"
	s.Write([]byte(headAndFoot))
	for _, line := range welcome {
		s.Write([]byte(`| ` + line + strings.Repeat(` `, maxLen-2-len(line)) + " |\n"))
	}
	s.Write([]byte(headAndFoot))

	// == wait ==

	if _, err := pty.Wait(); err != nil {
		s.Exit(1)
	} else {
		s.Exit(0)
	}
}
