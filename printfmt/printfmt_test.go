package printfmt_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/printfmt"
)

func TestBasic(t *testing.T) {

	w := strings.Builder{}

	o := printfmt.New(&w, 2)

	o.F("line 1")
	o.F(" extra")
	o.F(" extra\n")
	o.F("line 2")
	o.F(" extra")
	o.F(" extra\n")
	o.FinishLine() // should be nop
	o.F("line 3")
	o.F(" extra extra \nline 4 extra extra\nline 5 extra")
	o.F(" extra")
	o.FinishLine()

	fmt.Println("out: ", w.String())

	require.Equal(t, ("" +
		"    line 1 extra extra\n" +
		"    line 2 extra extra\n" +
		"    line 3 extra extra \n" +
		"    line 4 extra extra\n" +
		"    line 5 extra extra\n"), w.String())
}
