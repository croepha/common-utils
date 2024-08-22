package regexbuilder_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/regexbuilder"
)

var B = regexbuilder.B

var numberIdx = -1

var exampleRegex = B.A(
	B.R(`^`).Q(`match this exactly.***!!!`).R(`$`),
	B.Q(`captureThisNumber:`).GG(&numberIdx, `[[:digit:]]+`),
).Compile()

func TestBasic(t *testing.T) {

	require.Equal(t, false, exampleRegex.MatchString(`no match`))
	require.Equal(t, true, exampleRegex.MatchString(`match this exactly.***!!!`))

	m0 := exampleRegex.FindStringSubmatch("qwerq werq werqw er captureThisNumber:1234 qwqwerqwer qwe qwerqwe q qwer")
	require.Equal(t, "1234", m0[1])

}
