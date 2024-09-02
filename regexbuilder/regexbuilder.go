package regexbuilder

import (
	"fmt"
	"regexp"
	"slices"
	"sync/atomic"
)

/*
This is an experiment in some go code syntax sugar for regular expression construction
*/

type PatternBuilder struct {
	partialPattern string
	groups         []addedGroup
}

// TODO: Calling code can mutate this, probably should have a function return constructor instead
var B = PatternBuilder{}

/*
TODO this bind thing didn't work out as well as I was hoping

Maybe we can figure out how to bind this to a struct somehow?
or maybe we return the globally unique name through the out parameter?
or maybe this?:

B.G(..., func(user any, bind int) { user.(myStruct).group0Bind = bind; })

or this?:

B.G(..., func(user any, submatch []byte ) { user.(myStruct).group0 = submatch; })

and the functions would get called on match.  Note: unfortunately, this would still only
 call the function once, even if there were multiple matches for this group

*/

// Adds a group, idxBind is set to the group index on Compile
func (b PatternBuilder) G(idxBind *int, pattern PatternBuilder) PatternBuilder {
	pattern.DebugCheck()
	if idxBind == nil {
		panic("idxBind")
	}

	patName := fmt.Sprintf("autoPaternName%06d", nextAutoGroupKeyIDX.Add(1)-1)
	b.partialPattern += `(?P<` + patName + `>` + pattern.partialPattern + `)`

	b = b.addGroups(addedGroup{idxBind, patName})
	b = b.addGroups(pattern.groups...)
	return b
}

var nextAutoGroupKeyIDX atomic.Uint64

// Like G, but you can supply a pattern string instead of a struct
func (b PatternBuilder) GG(idxBind *int, pattern string) PatternBuilder {
	return b.G(idxBind, B.R(pattern))
}

type addedGroup struct {
	idxBind    *int
	paternName string
}

func (b PatternBuilder) addGroups(groups ...addedGroup) PatternBuilder {
	b.groups = slices.Concat(b.groups, groups)
	return b
}

// panics if the current pattern doesn't compile
func (b PatternBuilder) DebugCheck() PatternBuilder {
	// TODO: Perhaps use build tags to make this a NOP for release builds?
	regexp.MustCompile(b.partialPattern)
	return b
}

// Appends raw regex pattern
func (b PatternBuilder) R(rawPattern string) PatternBuilder {
	b.partialPattern += rawPattern
	return b
}

// Appends annother builder to this one
func (b PatternBuilder) B(source PatternBuilder) PatternBuilder {
	b.partialPattern += source.partialPattern
	b = b.addGroups(source.groups...)
	return b
}

// Append to the pattern the literal string, (this escapes/quotes the string)
func (b PatternBuilder) Q(s string) PatternBuilder {
	pat := regexp.QuoteMeta(s)
	b.partialPattern += pat
	return b
}

// Append to the pattern the literal string, (this escapes/quotes the []byte)
func (b PatternBuilder) QQ(s []byte) PatternBuilder {
	return b.Q(string(s))
}

// Alternations, adds a set of patterns to match. ie `(?:alts0|alts1,altsn...)`
func (b PatternBuilder) A(alts ...PatternBuilder) PatternBuilder {
	b = b.R(`(?:`)
	for i, alt := range alts {
		if i != 0 {
			b = b.R(`|`)
		}
		b = b.B(alt)
	}
	b = b.R(`)`)
	return b
}

// Compiles, and initializes the group IDXs.  Panics on bad regex...
func (b PatternBuilder) Compile() *regexp.Regexp {
	rgx := regexp.MustCompile(b.partialPattern)
	for _, g := range b.groups {
		*g.idxBind = rgx.SubexpIndex(g.paternName)
	}
	return rgx
}

// REGEXP TODO: Add negative-pattern:
// this: (`[^a]*`) works if the closer is one character ("a"), for multiple chars ("abcd") you need do something like this (`(a[^b]|ab[^c]|abc[^d])*`)
// See reference: https://stackoverflow.com/a/26792316/693869

// REGEXP TODO: Methods for named character classes like [[:space:]]
// REGEXP TODO: Methods for many, range, (call it repeat?), should it take a group as an argument?
