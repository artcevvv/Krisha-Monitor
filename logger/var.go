package logger

import (
	"os"
	"strings"
)

type FileLogger struct {
	File        *os.File
	DebugMode   bool
	PrintErrors bool
	Replacer    *strings.Replacer
}
