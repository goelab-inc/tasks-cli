package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Path() (string, error) {
	if p := os.Getenv("TASK_FILE"); p != "" {
		return p, nil
	}
	h, e := os.UserHomeDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(h, ".config", "task", "tasks.json"), nil
}
func Load(path string) ([]Item, error) {
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return []Item{}, nil
	}
	if e != nil {
		return nil, e
	}
	var items []Item
	if e = json.Unmarshal(b, &items); e != nil {
		return nil, fmt.Errorf("read task file: %w", e)
	}
	return items, nil
}
func Save(path string, items []Item) error {
	b, e := json.MarshalIndent(items, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	tmp := path + ".tmp"
	if e = os.WriteFile(tmp, append(b, '\n'), 0600); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
