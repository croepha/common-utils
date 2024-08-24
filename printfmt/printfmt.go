package printfmt

import (
	"fmt"
	"io"
	"strings"
)

type fmtWrapper struct {
	w           io.Writer
	indentLevel int
	lineStarted bool
}

func New(writer io.Writer, indentLevel int) fmtWrapper {
	return fmtWrapper{
		w:           writer,
		indentLevel: indentLevel,
	}
}

// TODO: Buffering?  use with defer fmt.Flush()?
// TODO: Automatically add space to seperate fields? Or actually maybe an option to trim redundant space?
// TODO: Custom prefix?

// TODO: There may also be other interesting stuff we could do with output
// like printing tables, color formatting...

// TODO: Generally not checking errors on output write.
//   We could check for errors, and perhaps panic, but generally it's
//   assumed that write will not error

func (f *fmtWrapper) F(format string, a ...any) {
	prefix := strings.Repeat("  ", f.indentLevel)

	// TODO: probably should implrement a writer for this for perf
	s := fmt.Sprintf(format, a...)
	ss := strings.Split(s, "\n")

	for si := range ss {
		p := prefix
		if f.lineStarted {
			p = ""
		}

		f.lineStarted = false
		end := "\n"
		if si == len(ss)-1 {
			end = ""
			if len(ss[si]) == 0 {
				break
			}
			f.lineStarted = true
		}
		fmt.Fprintf(f.w, "%s%s%s", p, ss[si], end)
	}
}
func (f *fmtWrapper) FinishLine() {
	if f.lineStarted {
		fmt.Fprintf(f.w, "\n")
		f.lineStarted = false
	}
}
