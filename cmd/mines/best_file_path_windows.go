//+build windows

package main

import (
	"os"
	"path/filepath"
)

func bestFilePath() string {
	return filepath.Join(os.Getenv("APPDATA"), "mines_best_times")
}
