package glob

import (
	"strings"
)

// GlobExpr represents a compiled glob expression.
type GlobExpr struct {
	parts []string
	lead  bool
	trail bool
}

// Match will attempt to match supplied string against receiving glob expression.
func (expr *GlobExpr) Match(s string) bool {
	// Empty pattern only matches empty
	if len(expr.parts) < 1 {
		// Pattern is empty,
		// -> 's' MUST be empty
		if !expr.lead {
			return len(s) < 1
		}

		// Pattern is glob,
		// -> matches all
		return true
	}

	cur := 0
	parts := expr.parts

	if !expr.lead {
		// No leading glob in pattern

		// 's' MUST start with first part
		idx := strings.Index(s, parts[0])
		if idx != 0 {
			return false
		}

		// Set next cursor pos, skip first
		cur = idx + len(parts[0])
		parts = parts[1:]
	}

	for _, part := range parts {
		// Look for start of next section
		idx := strings.Index(s[cur:], part)
		if idx < 0 {
			return false
		}

		// Set next cursor pos
		cur = idx + len(part)
	}

	if !expr.trail {
		// No trailing glob in pattern

		// MUST have reached end
		return (cur == len(s))
	}

	return true
}

// Match will attempt to match supplied string against supplied glob pattern.
func Match(pattern string, s string) bool {
	return Compile(pattern).Match(s)
}

// Compile will compile supplied glob pattern into a GlobExpr.
func Compile(pattern string) *GlobExpr {
	var lead, trail bool
	var parts []string

	switch len(pattern) {
	// No pattern
	case 0:

	// Either single char, or
	// single glob (matches all)
	case 1:
		lead = (pattern[0] == '*')

	// Pattern needs decoding
	default:
		b := []rune{}
		delim := false

		// Iterate the pattern for globs
		for i, r := range pattern {
			if r == '*' {
				if delim {
					// Delimited, add as char
					b = append(b, '*')
					delim = false
					continue
				}

				if i == 0 {
					// Leading glob
					lead = true
				} else if i == len(pattern)-1 {
					// Trailing glob
					trail = true
				}

				// Append the next pattern part
				parts = append(parts, string(b))
				b = b[:0]
			} else if r == '\\' && !delim {
				// Toggle delim status
				delim = true
			} else {
				if delim {
					// Delim was char
					b = append(b, '\\')
					delim = false
				}

				// Write to part builder
				b = append(b, r)
			}
		}

		// Append remaining part
		if len(b) > 0 {
			parts = append(parts, string(b))
		}
	}

	return &GlobExpr{
		parts: parts,
		lead:  lead,
		trail: trail,
	}
}
