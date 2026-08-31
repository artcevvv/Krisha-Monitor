package logger

import "fmt"

func (l *FileLogger) Debug(args ...interface{}) {
	if l.DebugMode {
		fmt.Fprint(l.File, "[DEBUG] ")
		fmt.Fprintln(l.File, args...)
	}
}

func (l *FileLogger) Debugf(format string, args ...interface{}) {
	if l.DebugMode {
		fmt.Fprint(l.File, "[DEBUG] ")
		if l.Replacer != nil {
			format = l.Replacer.Replace(format)
		}
		fmt.Fprintf(l.File, format+"\n", args...)
	}
}

func (l *FileLogger) Error(args ...interface{}) {
	if l.PrintErrors {
		fmt.Fprint(l.File, "[ERROR] ")
		fmt.Fprintln(l.File, args...)
	}
}

func (l *FileLogger) Errorf(format string, args ...interface{}) {
	if l.PrintErrors {
		fmt.Fprint(l.File, "[ERROR] ")
		if l.Replacer != nil {
			format = l.Replacer.Replace(format)
		}
		fmt.Fprintf(l.File, format+"\n", args...)
	}
}
