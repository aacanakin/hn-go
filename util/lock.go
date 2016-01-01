package util

import (
	"fmt"
	"os"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func CleanLocks() error {

	var err error
	err = CleanLock("top")
	err = CleanLock("new")
	err = CleanLock("ask")
	err = CleanLock("jobs")
	err = CleanLock("show")
	return err
}

func CleanLock(storyType string) error {

	err := os.Remove(fmt.Sprintf("./storage/locks/%s.lock", storyType))
	return err
}
