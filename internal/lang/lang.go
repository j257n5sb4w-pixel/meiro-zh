// Package lang shows Meiro's interface in the language the user picked.
//
// Text is translated where it is drawn, not where it is set: the same notice,
// error or page title is written once and read back translated wherever it
// appears, and a message built at runtime out of a prefix and a name, such as
// "Added “x” to your library", is translated by its pieces.
package lang

import (
	"os"
	"strings"

	"github.com/egoist/mygo"
)

// The languages the interface can show.
const (
	// Auto follows the language of the system.
	Auto = "auto"
	// English shows the interface as it is written.
	English = "en"
	// Chinese shows the interface in Simplified Chinese.
	Chinese = "zh"
)

// name is the tag of the language on show, English or Chinese.
var name = English

// Set picks the language the interface shows. It takes Auto, English or
// Chinese, and anything else means Auto. Auto reads the languages the system
// prefers, and answers English when Meiro is not running with a window yet, as
// in a test.
func Set(language string) {
	switch language {
	case Chinese:
		name = Chinese
	case English:
		name = English
	default:
		name = detect()
	}
}

// Current is the tag of the language on show.
func Current() string { return name }

// Shown reports whether the interface shows Chinese.
func Shown() bool { return name == Chinese }

// detect is the language a fresh install starts in: the system's, when it is
// Chinese, and English otherwise.
func detect() string {
	for _, entry := range []string{"MYGO_LOCALE", "LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if strings.HasPrefix(strings.ToLower(os.Getenv(entry)), "zh") {
			return Chinese
		}
	}
	if !mygo.App.IsReady() {
		return English
	}
	for _, tag := range mygo.App.PreferredLanguages() {
		if strings.HasPrefix(strings.ToLower(tag), "zh") {
			return Chinese
		}
	}
	return English
}

// HL is the language YouTube Music answers in, as hl takes it, so that the
// titles it sends back are in the language the interface shows.
func HL() string {
	if name == Chinese {
		return "zh-CN"
	}
	return "en"
}

// T is the text to show for s.
func T(s string) string {
	if name != Chinese || s == "" {
		return s
	}
	return translate(s)
}
