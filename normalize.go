package thaiaddress

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// thaiPrefixes are leading Thai administrative-area prefixes stripped by
// NormalizeName so that, for example, "อำเภอเมืองเชียงใหม่" and "เมืองเชียงใหม่"
// (or "เขตพระนคร" and "พระนคร") compare equal. Longer prefixes are listed first
// so the most specific one is matched.
var thaiPrefixes = []string{
	"กิ่งอำเภอ",
	"จังหวัด",
	"อำเภอ",
	"ตำบล",
	"แขวง",
	"เขต",
	"อ.",
	"ต.",
}

// englishPrefixes are leading English (RTGS) administrative-area prefixes
// stripped by NormalizeName. They are compared case-insensitively against the
// already-lowercased string, so they are listed in lowercase here. Each carries
// its trailing space so it only matches when used as a real prefix word.
var englishPrefixes = []string{
	"changwat ",
	"amphoe ",
	"khwaeng ",
	"tambon ",
	"khet ",
}

// NormalizeName canonicalizes an administrative-area name for matching. It:
//   - applies Unicode NFC normalization,
//   - trims surrounding whitespace and collapses internal runs of whitespace to
//     a single ASCII space,
//   - lowercases the string (affecting only the ASCII/Latin parts; Thai has no
//     case), and
//   - strips a single leading Thai or English administrative prefix
//     ("จังหวัด"/"Changwat ", "อำเภอ"/"อ."/"Amphoe ", "เขต"/"Khet ",
//     "ตำบล"/"ต."/"Tambon ", "แขวง"/"Khwaeng ", "กิ่งอำเภอ").
//
// A prefix is never stripped when it constitutes the entire remaining name, so
// the district name "เมือง" is preserved rather than reduced to "".
func NormalizeName(s string) string {
	// NFC first so that composed/decomposed forms compare equal and prefix
	// matching sees canonical code points.
	s = norm.NFC.String(s)

	// Collapse whitespace: strings.Fields splits on any Unicode space and drops
	// empties, which also trims the ends.
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return ""
	}

	// Lowercase for case-insensitive EN matching; harmless for Thai.
	s = strings.ToLower(s)

	s = stripPrefix(s)
	return s
}

// stripPrefix removes a single leading admin prefix from an already
// NFC/lowercased/whitespace-collapsed string, unless doing so would empty it.
func stripPrefix(s string) string {
	// English prefixes already include a trailing space, so a non-empty
	// remainder is guaranteed when they match.
	for _, p := range englishPrefixes {
		if strings.HasPrefix(s, p) {
			rest := strings.TrimSpace(s[len(p):])
			if rest != "" {
				return rest
			}
			return s
		}
	}
	for _, p := range thaiPrefixes {
		if strings.HasPrefix(s, p) && len(s) > len(p) {
			rest := strings.TrimSpace(s[len(p):])
			if rest != "" {
				return rest
			}
		}
	}
	return s
}
