package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SaveAppStartTime(dir string) error {
	t := fmt.Sprintf("%v\n", time.Now())
	err := os.MkdirAll(filepath.Dir(dir), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "restart.txt"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(t); err != nil {
		return err
	}
	return nil
}

func IsFileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
