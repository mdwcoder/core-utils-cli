package output

import (
	"fmt"
	"os"
	"strings"
)

func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func Println(a ...interface{}) {
	fmt.Println(a...)
}

func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func Errorln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func Status(label string, ok bool, msg string) {
	if ok {
		fmt.Printf("  %s  %s\n", green("OK"), label)
	} else {
		fmt.Printf("  %s  %s: %s\n", red("ERROR"), label, msg)
	}
}

func StatusWarn(label string, msg string) {
	fmt.Printf("  %s  %s: %s\n", yellow("WARN"), label, msg)
}

func green(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

func red(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[31m" + s + "\033[0m"
}

func yellow(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

func Bold(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func HasColor() bool {
	return os.Getenv("NO_COLOR") == "" && !isTermDumb()
}

func isTermDumb() bool {
	return strings.ToLower(os.Getenv("TERM")) == "dumb"
}
