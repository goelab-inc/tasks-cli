package task

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Item struct {
	ID           string        `json:"id"`
	Text         string        `json:"text"`
	Priority     int           `json:"priority"` // 1 low, 2 medium, 3 high
	Project      string        `json:"project,omitempty"`
	Estimate     time.Duration `json:"estimate,omitempty"`
	Due          *time.Time    `json:"due,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	Status       string        `json:"status"`
	SnoozedUntil *time.Time    `json:"snoozed_until,omitempty"`
}

// Parse converts quick-add tokens into task attributes, leaving ordinary words as task text.
func Parse(input string, now time.Time) (Item, error) {
	var t Item
	t.Priority = 2
	words := strings.Fields(input)
	var text []string
	for i := 0; i < len(words); i++ {
		w := words[i]
		switch {
		case strings.HasPrefix(strings.ToLower(w), "!"):
			switch strings.ToLower(strings.TrimPrefix(w, "!")) {
			case "h", "high", "3":
				t.Priority = 3
			case "m", "med", "medium", "2":
				t.Priority = 2
			case "l", "low", "1":
				t.Priority = 1
			default:
				text = append(text, w)
			}
		case strings.HasPrefix(w, "@") && len(w) > 1:
			if t.Project != "" {
				text = append(text, w)
			} else {
				t.Project = strings.TrimPrefix(w, "@")
			}
		case strings.HasPrefix(w, "+"):
			d, err := time.ParseDuration(strings.TrimPrefix(w, "+"))
			if err == nil && d > 0 {
				t.Estimate = d
			} else {
				text = append(text, w)
			}
		case strings.HasPrefix(strings.ToLower(w), "due:"):
			due, err := ParseDue(w[len("due:"):], now)
			if err != nil {
				return Item{}, err
			}
			t.Due = &due
		default:
			text = append(text, w)
		}
	}
	t.Text = strings.TrimSpace(strings.Join(text, " "))
	if t.Text == "" {
		return Item{}, fmt.Errorf("task text is required")
	}
	t.CreatedAt = now
	t.Status = "backlog"
	t.ID, _ = NewID()
	return t, nil
}

func ParseDue(s string, now time.Time) (time.Time, error) {
	loc := now.Location()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch strings.ToLower(s) {
	case "today":
		return day, nil
	case "tomorrow":
		return day.AddDate(0, 0, 1), nil
	}
	weekdays := map[string]time.Weekday{"sun": time.Sunday, "sunday": time.Sunday, "mon": time.Monday, "monday": time.Monday, "tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday, "wed": time.Wednesday, "wednesday": time.Wednesday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday, "fri": time.Friday, "friday": time.Friday, "sat": time.Saturday, "saturday": time.Saturday}
	if wd, ok := weekdays[strings.ToLower(s)]; ok {
		delta := (int(wd) - int(day.Weekday()) + 7) % 7
		if delta == 0 {
			delta = 7
		}
		return day.AddDate(0, 0, delta), nil
	}
	d, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date %q (use today, tomorrow, weekday, or YYYY-MM-DD)", s)
	}
	return d, nil
}

func NewID() (string, error) {
	const alphabet = "abcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, 5)
	for i := range b {
		n, e := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if e != nil {
			return "", e
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
