package view

import (
	"fmt"
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format("2006/01/02 15:04")
}

func StatusClass(status string) string {
	switch status {
	case "Receiving":
		return "success"
	case "Idle":
		return "warning"
	case "Error":
		return "danger"
	default:
		return "secondary"
	}
}

func UptimeFmt(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%d:%02d", h, m)
}
