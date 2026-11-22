package task

import "time"

// Score uses priority (10/20/30), urgency (overdue +50; otherwise up to +35 within 7 days),
// backlog aging (+0.5/day, capped at 14), in-progress (+15), and quick-win (+8 <=30m, +4 <=2h).
func Score(t Item, now time.Time) float64 {
	s := float64(t.Priority * 10)
	if t.Due != nil {
		days := t.Due.Sub(now).Hours() / 24
		if days < 0 {
			s += 50
		} else if days <= 7 {
			s += 35 * (1 - days/7)
		}
	}
	age := now.Sub(t.CreatedAt).Hours() / 24
	if age > 0 {
		if age > 28 {
			age = 28
		}
		s += age * 0.5
	}
	if t.Status == "in-progress" {
		s += 15
	}
	if t.Estimate > 0 {
		if t.Estimate <= 30*time.Minute {
			s += 8
		} else if t.Estimate <= 2*time.Hour {
			s += 4
		}
	}
	return s
}
