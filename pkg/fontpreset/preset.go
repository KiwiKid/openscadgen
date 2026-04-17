// Package fontpreset expands hardcoded OpenSCAD text() font strings (no fc-list / fontconfig queries).
package fontpreset

import (
	"sort"
	"strings"
)

// HardcodedFontFaces is the canonical list: either "Family:style=Style" or a bare family name.
// Edit this slice when you want more coverage.
var HardcodedFontFaces = []string{
	// macOS (OpenSCAD font list samples)
	".Hiragino Kaku Gothic Interface:style=W0",
	".Hiragino Kaku Gothic Interface:style=W1",
	".Hiragino Kaku Gothic Interface:style=W2",
	".Hiragino Kaku Gothic Interface:style=W4",
	".Hiragino Kaku Gothic Interface:style=W5",
	".Hiragino Kaku Gothic Interface:style=W6",
	".Hiragino Kaku Gothic Interface:style=W7",
	".Hiragino Kaku Gothic Interface:style=W8",
	".Hiragino Kaku Gothic Interface:style=W9",
	".Hiragino Sans GB Interface:style=W3",
	".Hiragino Sans GB Interface:style=W6",
	".Keyboard:style=Regular",
	".KufiStandardGK PUA:style=Regular",
	".LastResort:style=Regular",
	".Lucida Grande UI:style=Bold",
	".Lucida Grande UI:style=Regular",
	".Muna PUA:style=Black",
	".Muna PUA:style=Bold",
	".Muna PUA:style=Regular",
	".Nadeem PUA:style=Regular",
	".New York:style=Black",
	".New York:style=Bold",
	".New York:style=Regular Italic",
	".New York:style=Semibold Italic",
	// OpenSCAD built-ins (all platforms)
	"Liberation Mono",
	"Liberation Sans",
	"Liberation Serif",
}

// ExpandSentinelsInPlace replaces known sentinel strings with []interface{}. Mutates m.
func ExpandSentinelsInPlace(m map[string]interface{}) error {
	if m == nil {
		return nil
	}
	for k, v := range m {
		s, ok := v.(string)
		if !ok {
			continue
		}
		switch s {
		case SentinelHardcodedFaces, SentinelFontconfigFaces:
			m[k] = stringsToInterfaceSlice(HardcodedFontFaces)
		case SentinelFontconfigFamilies:
			m[k] = stringsToInterfaceSlice(uniqueFamiliesFromFaces(HardcodedFontFaces))
		}
	}
	return nil
}

func stringsToInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func uniqueFamiliesFromFaces(faces []string) []string {
	seen := make(map[string]struct{})
	for _, f := range faces {
		fam := f
		if i := strings.Index(f, ":style="); i >= 0 {
			fam = f[:i]
		}
		fam = strings.TrimSpace(fam)
		if fam == "" {
			continue
		}
		seen[fam] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for f := range seen {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}
