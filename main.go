package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/goelab-inc/tasks-cli/internal/task"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "task:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	path, e := task.Path()
	if e != nil {
		return e
	}
	items, e := task.Load(path)
	if e != nil {
		return e
	}
	now := time.Now()
	if len(args) == 0 {
		return interactiveAdd(path, items, now)
	}
	cmd := args[0]
	if cmd == "next" {
		cmd = "catchup"
	}
	switch cmd {
	case "add":
		if len(args) < 2 {
			return interactiveAdd(path, items, now)
		}
		x, e := task.Parse(strings.Join(args[1:], " "), now)
		if e != nil {
			return e
		}
		items = append(items, x)
		if e = task.Save(path, items); e != nil {
			return e
		}
		fmt.Printf("Added %s: %s\n", color(x.ID, "36"), x.Text)
		return nil
	case "ls":
		all := false
		project := ""
		for _, a := range args[1:] {
			if a == "--all" {
				all = true
			} else if strings.HasPrefix(a, "@") {
				project = strings.TrimPrefix(a, "@")
			} else {
				return fmt.Errorf("unknown ls option %q", a)
			}
		}
		visible := make([]task.Item, 0)
		for _, x := range items {
			if x.Status == "done" && !all {
				continue
			}
			if project != "" && x.Project != project {
				continue
			}
			visible = append(visible, x)
		}
		sort.SliceStable(visible, func(i, j int) bool { return task.Score(visible[i], now) > task.Score(visible[j], now) })
		if len(visible) == 0 {
			fmt.Println("No tasks.")
			return nil
		}
		for _, x := range visible {
			printTask(x, now)
		}
		return nil
	case "catchup":
		ranked := rank(items, now)
		if len(ranked) == 0 {
			fmt.Println("Nothing to work on. Add one with: task add <what to do>")
			return nil
		}
		fmt.Println(color("NOW", "1;32") + "  " + ranked[0].Text + "  " + color("["+ranked[0].ID+"]", "36"))
		fmt.Println("    " + reason(ranked[0], now))
		for i := 1; i < len(ranked) && i < 3; i++ {
			fmt.Printf("NEXT  %s  [%s]\n", ranked[i].Text, ranked[i].ID)
		}
		return nil
	case "done", "start", "rm", "bump", "show", "edit", "snooze":
		if len(args) < 2 {
			return fmt.Errorf("usage: task %s <id>", cmd)
		}
		idx := find(items, args[1])
		if idx < 0 {
			return fmt.Errorf("no task matches id %q", args[1])
		}
		x := &items[idx]
		switch cmd {
		case "done":
			x.Status = "done"
			x.SnoozedUntil = nil
		case "start":
			x.Status = "in-progress"
			x.SnoozedUntil = nil
		case "bump":
			if x.Priority < 3 {
				x.Priority++
			}
		case "rm":
			items = append(items[:idx], items[idx+1:]...)
			if e = task.Save(path, items); e != nil {
				return e
			}
			fmt.Println("Removed task", args[1])
			return nil
		case "show":
			show(*x, now)
			return nil
		case "edit":
			if len(args) < 3 {
				return fmt.Errorf("usage: task edit <id> <inline syntax>")
			}
			parsed, err := task.Parse(strings.Join(args[2:], " "), now)
			if err != nil {
				return err
			}
			x.Text = parsed.Text
			x.Priority = parsed.Priority
			x.Project = parsed.Project
			x.Estimate = parsed.Estimate
			x.Due = parsed.Due
		case "snooze":
			if len(args) < 3 {
				return fmt.Errorf("usage: task snooze <id> <duration|date>")
			}
			when, err := snoozeTime(args[2], now)
			if err != nil {
				return err
			}
			x.SnoozedUntil = &when
		}
		if e = task.Save(path, items); e != nil {
			return e
		}
		fmt.Printf("Updated %s\n", x.ID)
		return nil
	default:
		return fmt.Errorf("unknown command %q (try add, ls, catchup, done, start, edit, rm, bump, snooze, show)", cmd)
	}
}
func interactiveAdd(path string, items []task.Item, now time.Time) error {
	fmt.Print("Task: ")
	s, e := bufio.NewReader(os.Stdin).ReadString('\n')
	if e != nil && len(s) == 0 {
		return e
	}
	x, e := task.Parse(strings.TrimSpace(s), now)
	if e != nil {
		return e
	}
	items = append(items, x)
	if e = task.Save(path, items); e != nil {
		return e
	}
	fmt.Printf("Added %s: %s\n", x.ID, x.Text)
	return nil
}
func rank(items []task.Item, now time.Time) []task.Item {
	r := []task.Item{}
	for _, x := range items {
		if x.Status == "done" || (x.SnoozedUntil != nil && x.SnoozedUntil.After(now)) {
			continue
		}
		r = append(r, x)
	}
	sort.SliceStable(r, func(i, j int) bool { return task.Score(r[i], now) > task.Score(r[j], now) })
	return r
}
func find(items []task.Item, id string) int {
	found := -1
	for i, x := range items {
		if x.ID == id || strings.HasPrefix(x.ID, id) {
			if found >= 0 {
				return -2
			}
			found = i
		}
	}
	return found
}
func printTask(x task.Item, now time.Time) {
	status := ""
	if x.Status == "in-progress" {
		status = color("in-progress", "33")
	}
	due := ""
	if x.Due != nil {
		due = " due:" + x.Due.Format("2006-01-02")
	}
	proj := ""
	if x.Project != "" {
		proj = " @" + x.Project
	}
	fmt.Printf("%-5s %-12s %s%s%s  (%.0f)\n", color(x.ID, "36"), status, x.Text, proj, due, task.Score(x, now))
}
func show(x task.Item, now time.Time) {
	printTask(x, now)
	fmt.Printf("Priority: %d | Estimate: %s | Created: %s\n", x.Priority, x.Estimate, x.CreatedAt.Format("2006-01-02"))
	if x.SnoozedUntil != nil {
		fmt.Printf("Snoozed until: %s\n", x.SnoozedUntil.Format(time.RFC3339))
	}
}
func reason(x task.Item, now time.Time) string {
	var why []string
	switch x.Priority {
	case 3:
		why = append(why, "high priority")
	case 1:
		why = append(why, "low priority")
	default:
		why = append(why, "medium priority")
	}
	if x.Due != nil {
		if x.Due.Before(now) {
			why = append(why, "overdue")
		} else if x.Due.Sub(now) <= 7*24*time.Hour {
			why = append(why, "due soon")
		}
	}
	if x.Status == "in-progress" {
		why = append(why, "already in progress")
	}
	if x.Estimate > 0 && x.Estimate <= 30*time.Minute {
		why = append(why, "quick win")
	}
	return "Why: " + strings.Join(why, ", ")
}
func snoozeTime(s string, now time.Time) (time.Time, error) {
	if strings.HasPrefix(s, "+") {
		d, e := time.ParseDuration(strings.TrimPrefix(s, "+"))
		if e != nil || d <= 0 {
			return time.Time{}, fmt.Errorf("invalid duration %q", s)
		}
		return now.Add(d), nil
	}
	return task.ParseDue(s, now)
}
func color(s, code string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	f, e := os.Stdout.Stat()
	if e != nil || f.Mode()&os.ModeCharDevice == 0 {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}
