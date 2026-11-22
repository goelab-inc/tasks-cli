package task

import (
	"testing"
	"time"
)

func TestParseInline(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	x, e := Parse("fix login bug !high @api +2h due:tomorrow", now)
	if e != nil {
		t.Fatal(e)
	}
	if x.Text != "fix login bug" || x.Priority != 3 || x.Project != "api" || x.Estimate != 2*time.Hour || x.Due == nil || x.Due.Day() != 23 {
		t.Fatalf("unexpected parse: %+v", x)
	}
}
func TestParseWeekdayAndBadDate(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	d, e := ParseDue("fri", now)
	if e != nil || d.Weekday() != time.Friday || d.Day() != 27 {
		t.Fatalf("due=%v err=%v", d, e)
	}
	if _, e = ParseDue("noday", now); e == nil {
		t.Fatal("expected invalid date")
	}
}
func TestScoreFactors(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	base := Item{Priority: 2, CreatedAt: now, Status: "backlog"}
	urgent := base
	due := now.Add(-time.Hour)
	urgent.Due = &due
	urgent.Status = "in-progress"
	urgent.Estimate = 20 * time.Minute
	if Score(urgent, now) <= Score(base, now) {
		t.Fatal("urgency, progress and quick-win should increase score")
	}
	old := base
	old.CreatedAt = now.Add(-14 * 24 * time.Hour)
	if Score(old, now) <= Score(base, now) {
		t.Fatal("aging should increase score")
	}
}
