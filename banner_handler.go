package gosshserver

import (
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
)

func bannerHandler(ctx ssh.Context) string {
	sb := strings.Builder{}
	//return "欢迎使用 SSH 服务" + "\n"

	// == welcome message ==

	welcome := []string{
		time.Now().Format(time.RFC3339),
		`Welcome ` + ctx.User() + ", " + ctx.RemoteAddr().String(),
	}
	var maxLen int
	for _, line := range welcome {
		maxLen = max(maxLen, len(line)+2)
	}
	headAndFoot := `+` + strings.Repeat(`-`, maxLen) + "+\n"
	sb.WriteString(headAndFoot)
	for _, line := range welcome {
		sb.WriteString(`| ` + line + strings.Repeat(` `, maxLen-2-len(line)) + " |\n")
	}
	sb.WriteString(headAndFoot)
	return sb.String()
}
