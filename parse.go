package thaiaddress

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ParseResult is the outcome of [ParseThaiAddress]: the administrative levels
// recognised in a free-text Thai address, plus whatever free text was left
// over.
//
// The parser is best-effort. Any of Province, District, Subdistrict and
// Postcode may be unset (nil pointers, or a zero Postcode) when the input did
// not pin that level unambiguously; the parser never guesses a level it cannot
// determine from the text. Whatever could not be attributed to an
// administrative level — typically the house number, หมู่/หมู่บ้าน, ถนน/ซอย —
// survives in Remainder.
type ParseResult struct {
	// Province is the recognised province (จังหวัด), or nil if none could be
	// pinned.
	Province *Province
	// District is the recognised district (อำเภอ/เขต), or nil if none could be
	// pinned.
	District *District
	// Subdistrict is the recognised subdistrict (ตำบล/แขวง), or nil if none
	// could be pinned.
	Subdistrict *Subdistrict
	// Postcode is the 5-digit postal code found in the text, or 0 if none.
	//
	// It is reported as found in the input. On contradictory input — a marked
	// area whose own postcode differs from the typed one — the parser trusts the
	// marked names for the admin chain and keeps this postcode unchanged, so it
	// may not equal the resolved area's postcode. The resolved
	// Province/District/Subdistrict chain itself is always internally
	// consistent.
	Postcode int
	// Remainder is the leftover free text (house number, road, village, …) after
	// the matched postcode and administrative-area tokens were removed, with
	// whitespace collapsed and trimmed. Empty when nothing is left over.
	Remainder string
}

// IsComplete reports whether all three administrative levels were resolved.
func (r ParseResult) IsComplete() bool {
	return r.Province != nil && r.District != nil && r.Subdistrict != nil
}

// IsEmpty reports whether nothing administrative was recognised at all (no
// province, district, subdistrict or postcode). Remainder may still be
// non-empty.
func (r ParseResult) IsEmpty() bool {
	return r.Province == nil && r.District == nil && r.Subdistrict == nil && r.Postcode == 0
}

// bangkokAliases stand in for "province กรุงเทพมหานคร" (code 10), longest first
// so the most text is consumed.
var bangkokAliases = []string{
	"กรุงเทพมหานคร",
	"กรุงเทพฯ",
	"กรุงเทพ",
	"กทม.",
	"กทม",
}

// provinceMarkers, longest first so "จังหวัด" wins over "จ.".
var provinceMarkers = []string{"จังหวัด", "จ."}

// districtMarkers, longest first.
var districtMarkers = []string{"อำเภอ", "เขต", "อ."}

// subdistrictMarkers, longest first.
var subdistrictMarkers = []string{"ตำบล", "แขวง", "ต."}

// allMarkers across every level, used to terminate a name run so that, e.g.,
// "ต.บางรักอ.บางรัก" splits at the second marker even without whitespace.
var allMarkers = []string{
	"จังหวัด",
	"อำเภอ",
	"ตำบล",
	"แขวง",
	"เขต",
	"จ.",
	"อ.",
	"ต.",
}

// span is a half-open byte range of the ORIGINAL input that has been consumed
// by a match and must be excised from the remainder.
type span struct {
	start int
	end   int
}

// arabicized holds the digit-normalised working string together with a map from
// each work-string byte offset back to the corresponding byte offset in the
// original input. Because Thai digits (3 UTF-8 bytes) become ASCII (1 byte),
// the two strings have different byte lengths, so spans found while scanning
// work must be translated back onto original via origOff before excision.
type arabicized struct {
	work string
	// origOff[i] is the original byte offset corresponding to work byte i, for
	// i in [0, len(work)]. The terminal entry maps len(work) -> len(original),
	// so end offsets translate correctly.
	origOff []int
}

// arabicizeDigits maps Thai digits ๐–๙ (U+0E50–U+0E59) onto Arabic 0–9,
// building the work string rune-by-rune and recording, for every byte of the
// result, the byte offset of the corresponding rune in the original input.
func arabicizeDigits(s string) arabicized {
	var b strings.Builder
	b.Grow(len(s))
	// origOff sized for the worst case (no shrinkage) plus the terminal entry.
	origOff := make([]int, 0, len(s)+1)
	for origIdx, r := range s {
		if r >= 0x0E50 && r <= 0x0E59 {
			b.WriteByte(byte('0' + (r - 0x0E50)))
			origOff = append(origOff, origIdx)
		} else {
			// Append the rune's own UTF-8 bytes; map each of them to origIdx (the
			// start of this rune in the original), which is what indexOf-style
			// span starts need.
			start := b.Len()
			b.WriteRune(r)
			for j := start; j < b.Len(); j++ {
				origOff = append(origOff, origIdx)
			}
		}
	}
	// Terminal entry: a work offset equal to len(work) maps to len(original).
	origOff = append(origOff, len(s))
	return arabicized{work: b.String(), origOff: origOff}
}

// toOrig translates a half-open work-byte span [start,end) into the
// corresponding original-byte span.
func (a arabicized) toOrig(start, end int) span {
	return span{start: a.origOff[start], end: a.origOff[end]}
}

// markerHit is a marker match: the resolved name text and the work-string span
// (marker + name) to excise.
type markerHit struct {
	name  string
	start int // work-byte offset of the marker
	end   int // work-byte offset just past the name
}

// ParseThaiAddress parses a free-text Thai address into its administrative
// levels.
//
// The parser combines two signals and reconciles them so they always agree:
//
//  1. Postcode — the first 5-digit run (Thai digits ๐–๙ are accepted and
//     normalised) that [ByPostcode] recognises. Its subdistricts form a
//     candidate set; the result is pinned only to the province/district that
//     all of those subdistricts share, because one postcode can span several
//     districts (and occasionally provinces).
//  2. Marked names — tokens introduced by an area marker: จังหวัด/จ. for the
//     province (plus the Bangkok aliases กทม/กทม./กรุงเทพ/กรุงเทพฯ/
//     กรุงเทพมหานคร, all → code 10), อำเภอ/อ./เขต for the district, and
//     ตำบล/ต./แขวง for the subdistrict. Each name is resolved with the package
//     name lookups. A bare name with no marker is not scanned for; it is left
//     in the remainder rather than guessed at.
//
// The signals are intersected: a subdistrict must belong to the chosen
// district, which must belong to the chosen province, and everything must be
// consistent with the postcode candidate set. Thai subdistrict (and district)
// names repeat across provinces; the duplicate is resolved using the postcode
// and the higher levels. If a level stays ambiguous it is left nil rather than
// guessed.
//
// The matched postcode and consumed area tokens are removed from the input; the
// collapsed, trimmed leftover (house number, หมู่/หมู่บ้าน, ถนน/ซอย, …) is
// returned in [ParseResult.Remainder].
//
// This function never panics: malformed or unrelated input yields an IsEmpty
// result whose Remainder is the cleaned input.
func ParseThaiAddress(input string) ParseResult {
	// Work on an Arabic-digit copy so postcode scanning is uniform; spans found
	// here are translated back onto the original via the byte-offset map.
	a := arabicizeDigits(input)
	work := a.work
	var spans []span

	// ---- 1. Postcode -------------------------------------------------------
	postcode := 0
	var postcodeCandidates []Subdistrict
	for _, m := range findDigitRuns(work) {
		value, _ := strconv.Atoi(work[m.start:m.end])
		subs := ByPostcode(value)
		if len(subs) > 0 {
			postcode = value
			postcodeCandidates = subs
			spans = append(spans, a.toOrig(m.start, m.end))
			break // first valid postcode wins
		}
	}

	// ---- 2. Marked names ---------------------------------------------------
	var markedDistricts []District
	var markedSubdistricts []Subdistrict
	var markedProvince *Province

	// Province via Bangkok aliases (longest alias first to consume the most).
	for _, alias := range bangkokAliases {
		if i := strings.Index(work, alias); i >= 0 {
			if p, ok := ProvinceByCode(10); ok {
				pp := p
				markedProvince = &pp
				spans = append(spans, a.toOrig(i, i+len(alias)))
			}
			break
		}
	}
	// Province via จังหวัด/จ. marker.
	if markedProvince == nil {
		if hit, ok := afterMarker(work, provinceMarkers); ok {
			if ps := FindProvinces(hit.name); len(ps) > 0 {
				pp := ps[0] // province names are unique
				markedProvince = &pp
				spans = append(spans, a.toOrig(hit.start, hit.end))
			}
		}
	}
	// District via อำเภอ/อ./เขต.
	if hit, ok := afterMarker(work, districtMarkers); ok {
		if ds := FindDistricts(hit.name); len(ds) > 0 {
			markedDistricts = append(markedDistricts, ds...)
			spans = append(spans, a.toOrig(hit.start, hit.end))
		}
	}
	// Subdistrict via ตำบล/ต./แขวง.
	if hit, ok := afterMarker(work, subdistrictMarkers); ok {
		if ss := FindSubdistricts(hit.name); len(ss) > 0 {
			markedSubdistricts = append(markedSubdistricts, ss...)
			spans = append(spans, a.toOrig(hit.start, hit.end))
		}
	}

	// ---- 3. Reconcile ------------------------------------------------------
	province := markedProvince
	var districtPool []District
	if len(markedDistricts) > 0 {
		districtPool = append(districtPool, markedDistricts...)
	}
	var subPool []Subdistrict
	if len(markedSubdistricts) > 0 {
		subPool = append(subPool, markedSubdistricts...)
	}

	// If we have a postcode, it constrains every level.
	if postcodeCandidates != nil {
		// Intersect subdistrict pool with the postcode's subdistricts.
		if len(subPool) > 0 {
			allowed := map[int]bool{}
			for _, s := range postcodeCandidates {
				allowed[s.Code] = true
			}
			var narrowed []Subdistrict
			for _, s := range subPool {
				if allowed[s.Code] {
					narrowed = append(narrowed, s)
				}
			}
			// Keep the narrowing only if it left something; otherwise the marked
			// subdistrict disagrees with the postcode — trust the marked name but
			// still keep the postcode for its own province/district pinning.
			if len(narrowed) > 0 {
				subPool = narrowed
			}
		} else {
			// No subdistrict named: the postcode's subdistricts are used only to
			// pin shared higher levels (not to choose a subdistrict).
			subPool = nil
		}

		// Districts represented by the postcode.
		pcDistricts := map[int]bool{}
		for _, s := range postcodeCandidates {
			pcDistricts[s.DistrictCode] = true
		}
		if len(districtPool) > 0 {
			var narrowed []District
			for _, d := range districtPool {
				if pcDistricts[d.Code] {
					narrowed = append(narrowed, d)
				}
			}
			if len(narrowed) > 0 {
				districtPool = narrowed
			}
		} else if len(pcDistricts) == 1 {
			// No district named, but every postcode candidate shares one
			// district: pin it.
			var only int
			for c := range pcDistricts {
				only = c
			}
			if d, ok := DistrictByCode(only); ok {
				districtPool = []District{d}
			}
		}

		// Pin province from the postcode if every candidate shares one and there
		// is no conflicting marked province.
		pcProvinces := map[int]bool{}
		for _, s := range postcodeCandidates {
			pcProvinces[s.DistrictCode/100] = true
		}
		if province == nil && len(pcProvinces) == 1 {
			var only int
			for c := range pcProvinces {
				only = c
			}
			if p, ok := ProvinceByCode(only); ok {
				pp := p
				province = &pp
			}
		}
	}

	// Narrow district pool by province context.
	if province != nil && len(districtPool) > 0 {
		var narrowed []District
		for _, d := range districtPool {
			if d.ProvinceCode == province.Code {
				narrowed = append(narrowed, d)
			}
		}
		if len(narrowed) > 0 {
			districtPool = narrowed
		}
	}
	// Narrow subdistrict pool by province context.
	if province != nil && len(subPool) > 0 {
		var narrowed []Subdistrict
		for _, s := range subPool {
			if s.DistrictCode/100 == province.Code {
				narrowed = append(narrowed, s)
			}
		}
		if len(narrowed) > 0 {
			subPool = narrowed
		}
	}
	// Narrow subdistrict pool by district context.
	if len(districtPool) == 1 && len(subPool) > 0 {
		dCode := districtPool[0].Code
		var narrowed []Subdistrict
		for _, s := range subPool {
			if s.DistrictCode == dCode {
				narrowed = append(narrowed, s)
			}
		}
		if len(narrowed) > 0 {
			subPool = narrowed
		}
	}

	// Choose the subdistrict if exactly one remains.
	var subdistrict *Subdistrict
	if len(subPool) == 1 {
		ss := subPool[0]
		subdistrict = &ss
	}

	// If a subdistrict was chosen, it implies its district and province; fill any
	// gaps and let the implied values win (they are strictly more specific).
	if subdistrict != nil {
		if len(districtPool) == 0 || len(districtPool) > 1 {
			if d, ok := DistrictByCode(subdistrict.DistrictCode); ok {
				districtPool = []District{d}
			}
		}
		if province == nil {
			if p, ok := ProvinceByCode(subdistrict.DistrictCode / 100); ok {
				pp := p
				province = &pp
			}
		}
	}

	// Choose the district if exactly one remains.
	var district *District
	if len(districtPool) == 1 {
		dd := districtPool[0]
		district = &dd
	}

	// District implies province; fill the gap.
	if district != nil && province == nil {
		if p, ok := ProvinceByCode(district.ProvinceCode); ok {
			pp := p
			province = &pp
		}
	}

	// Final consistency guard: drop any level that contradicts a more specific
	// one (can only happen when a marked name disagreed with the postcode).
	if district != nil && province != nil && district.ProvinceCode != province.Code {
		district = nil
	}
	if subdistrict != nil {
		if district != nil && subdistrict.DistrictCode != district.Code {
			subdistrict = nil
		} else if province != nil && subdistrict.DistrictCode/100 != province.Code {
			subdistrict = nil
		}
	}

	// ---- 4. Remainder ------------------------------------------------------
	remainder := stripSpans(input, spans)

	return ParseResult{
		Province:    province,
		District:    district,
		Subdistrict: subdistrict,
		Postcode:    postcode,
		Remainder:   remainder,
	}
}

// findDigitRuns returns the half-open byte spans of every maximal run of
// exactly 5 consecutive ASCII digits in work (the \d{5} equivalent). Like
// Dart's RegExp.allMatches the matches are non-overlapping and left-to-right;
// a longer digit run yields its leading 5-digit windows in 5-byte strides,
// which matches RegExp(r'\d{5}') behaviour.
func findDigitRuns(work string) []span {
	var out []span
	i := 0
	for i+5 <= len(work) {
		if isFiveDigits(work, i) {
			out = append(out, span{i, i + 5})
			i += 5
			continue
		}
		i++
	}
	return out
}

func isFiveDigits(s string, i int) bool {
	for j := i; j < i+5; j++ {
		if s[j] < '0' || s[j] > '9' {
			return false
		}
	}
	return true
}

// afterMarker finds the first occurrence of any marker in work and grabs the
// area name immediately following it. The name runs until a delimiter:
// whitespace, an ASCII digit, a comma, or the start of another known marker.
// Returns ok=false if no marker is present or the trailing name is empty.
func afterMarker(work string, markers []string) (markerHit, bool) {
	for _, marker := range markers {
		// A dotted abbreviation (จ./อ./ต.) is often written with a space before
		// the name (e.g. "ต. สุเทพ"); skip that gap so the name is still
		// captured. The long markers bind directly to the name.
		skipGap := strings.HasSuffix(marker, ".")
		from := 0
		for {
			i := strings.Index(work[from:], marker)
			if i < 0 {
				break
			}
			i += from
			nameStart := i + len(marker)
			if skipGap {
				for nameStart < len(work) && (work[nameStart] == ' ' || work[nameStart] == '\t') {
					nameStart++
				}
			}
			nameEnd := nameEndIndex(work, nameStart)
			name := strings.TrimSpace(work[nameStart:nameEnd])
			if name != "" {
				return markerHit{name: name, start: i, end: nameEnd}, true
			}
			from = i + len(marker)
		}
	}
	return markerHit{}, false
}

// nameEndIndex returns the byte index where a name starting at start ends: at
// the first whitespace, ASCII digit, comma, or the start of another known
// marker.
func nameEndIndex(work string, start int) int {
	i := start
	for i < len(work) {
		ch := work[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ',' ||
			(ch >= '0' && ch <= '9') {
			break
		}
		hitMarker := false
		for _, m := range allMarkers {
			// A marker only terminates the name if it does not start at `start`
			// itself (it is the boundary to the next token).
			if i > start && strings.HasPrefix(work[i:], m) {
				hitMarker = true
				break
			}
		}
		if hitMarker {
			break
		}
		// Advance by a full rune so we never split a multibyte Thai character.
		_, size := decodeRune(work[i:])
		i += size
	}
	return i
}

// decodeRune returns the first rune and the number of bytes it actually
// consumes, treating an empty input as a single byte so callers always make
// progress. It uses utf8.DecodeRuneInString so an invalid byte reports width 1
// (the bytes consumed), never the 3-byte width of the U+FFFD replacement rune —
// otherwise callers would over-advance past the string end on malformed UTF-8.
func decodeRune(s string) (rune, int) {
	if s == "" {
		return 0, 1
	}
	r, n := utf8.DecodeRuneInString(s)
	return r, n
}

// stripSpans removes the (original-byte) spans from original, then collapses
// whitespace and trims.
func stripSpans(original string, spans []span) string {
	if len(spans) == 0 {
		return strings.Join(strings.Fields(original), " ")
	}
	// Sort and merge overlapping spans, then keep the gaps between them.
	sorted := make([]span, len(spans))
	copy(sorted, spans)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start < sorted[j].start })
	var merged []span
	for _, s := range sorted {
		if len(merged) == 0 || s.start > merged[len(merged)-1].end {
			merged = append(merged, s)
		} else if s.end > merged[len(merged)-1].end {
			merged[len(merged)-1].end = s.end
		}
	}
	var b strings.Builder
	cursor := 0
	for _, s := range merged {
		if s.start > cursor {
			b.WriteString(original[cursor:s.start])
		}
		b.WriteByte(' ') // gap so adjacent tokens don't fuse
		cursor = s.end
	}
	if cursor < len(original) {
		b.WriteString(original[cursor:])
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
