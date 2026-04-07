package view

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/model"
)

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatTime":     formatTime,
		"formatDate":     formatDate,
		"formatDuration": formatDuration,
		"formatTimeRange": formatTimeRange,
		"uptimeFmt":      uptimeFmt,
		"statusClass":    statusClass,
		"autoLink":       autoLink,
		"nl2br":          nl2br,
		"isActive":       func(ci *model.ChannelInfo) bool { return ci.IsActive() },
		"add":            func(a, b int) int { return a + b },
		"sub":            func(a, b int) int { return a - b },
		"seq":            seq,
		"dict":           dict,
		"default":        defaultTo,
	}
}

func formatTime(t time.Time) string {
	return t.Format("2006/01/02 15:04")
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%d月%d日", t.Month(), t.Day())
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func formatTimeRange(start, end time.Time) string {
	if start.IsZero() {
		return ""
	}
	s := fmt.Sprintf("%d月%d日 %d時%02d分", start.Month(), start.Day(), start.Hour(), start.Minute())
	if !end.IsZero() {
		if start.Day() == end.Day() && start.Month() == end.Month() {
			s += fmt.Sprintf(" 〜 %d時%02d分", end.Hour(), end.Minute())
		} else {
			s += fmt.Sprintf(" 〜 %d月%d日 %d時%02d分", end.Month(), end.Day(), end.Hour(), end.Minute())
		}
	}
	return s
}

func uptimeFmt(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%d:%02d", h, m)
}

func statusClass(status string) string {
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

func autoLink(text string) template.HTML {
	words := strings.Fields(text)
	for i, w := range words {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			words[i] = fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener">%s</a>`, template.HTMLEscapeString(w), template.HTMLEscapeString(w))
		} else {
			words[i] = template.HTMLEscapeString(w)
		}
	}
	return template.HTML(strings.Join(words, " "))
}

func nl2br(text string) template.HTML {
	escaped := template.HTMLEscapeString(text)
	return template.HTML(strings.ReplaceAll(escaped, "\n", "<br>"))
}

func seq(start, end int) []int {
	var s []int
	for i := start; i <= end; i++ {
		s = append(s, i)
	}
	return s
}

func dict(pairs ...interface{}) map[string]interface{} {
	d := make(map[string]interface{})
	for i := 0; i+1 < len(pairs); i += 2 {
		key, _ := pairs[i].(string)
		d[key] = pairs[i+1]
	}
	return d
}

func defaultTo(def, val interface{}) interface{} {
	if val == nil || val == "" || val == 0 {
		return def
	}
	return val
}
