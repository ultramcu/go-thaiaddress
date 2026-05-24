package thaiaddress

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// Whitespace handling.
		{"trim", "  Bangkok  ", "bangkok"},
		{"collapse internal", "Chiang   Mai", "chiang mai"},
		{"tabs and newlines", "Nakhon\tSi\nThammarat", "nakhon si thammarat"},
		{"empty", "", ""},
		{"whitespace only", "   \t\n ", ""},

		// Lowercasing (EN only; Thai unaffected).
		{"lowercase en", "MUEANG CHIANG MAI", "mueang chiang mai"},

		// Thai prefix stripping.
		{"strip จังหวัด", "จังหวัดเชียงใหม่", "เชียงใหม่"},
		{"strip อำเภอ", "อำเภอเมืองเชียงใหม่", "เมืองเชียงใหม่"},
		{"strip อ.", "อ.จอมทอง", "จอมทอง"},
		{"strip ตำบล", "ตำบลในเมือง", "ในเมือง"},
		{"strip ต.", "ต.สุเทพ", "สุเทพ"},
		{"strip เขต", "เขตพระนคร", "พระนคร"},
		{"strip แขวง", "แขวงพระบรมมหาราชวัง", "พระบรมมหาราชวัง"},
		{"strip กิ่งอำเภอ", "กิ่งอำเภอทุ่งช้าง", "ทุ่งช้าง"},

		// English prefix stripping (case-insensitive via prior lowercasing).
		{"strip Changwat", "Changwat Chiang Mai", "chiang mai"},
		{"strip Amphoe", "Amphoe Mueang Chiang Mai", "mueang chiang mai"},
		{"strip Khet", "Khet Phra Nakhon", "phra nakhon"},
		{"strip Tambon", "Tambon Nai Mueang", "nai mueang"},
		{"strip Khwaeng", "Khwaeng Phra Borom Maha Ratchawang", "phra borom maha ratchawang"},
		{"strip Amphoe mixed case", "aMpHoE Chom Thong", "chom thong"},

		// The key equivalences the spec calls out.
		{"amphoe-prefixed equals bare TH", "อำเภอเมืองเชียงใหม่", NormalizeName("เมืองเชียงใหม่")},
		{"khet-prefixed equals bare TH", "เขตพระนคร", NormalizeName("พระนคร")},

		// Do NOT strip a prefix that is the whole string.
		{"keep bare เมือง (not a prefix in list)", "เมือง", "เมือง"},
		{"keep bare เขต", "เขต", "เขต"},
		{"keep bare อำเภอ", "อำเภอ", "อำเภอ"},
		{"keep bare ตำบล", "ตำบล", "ตำบล"},
		// English bare prefix without a trailing word never matches (no space).
		{"keep bare Khet (no trailing word)", "Khet", "khet"},
		{"keep bare Amphoe", "Amphoe", "amphoe"},

		// Only one leading prefix is stripped.
		{"single strip only", "ตำบลตำบล", "ตำบล"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeName(tt.in); got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalizeNFC verifies that decomposed and composed Unicode forms collapse
// to the same canonical string. "é" can be written as a single precomposed code
// point (U+00E9) or as "e" + combining acute accent (U+0065 U+0301); NFC must
// fold them together so a name typed either way compares equal.
func TestNormalizeNFC(t *testing.T) {
	// Build the two forms from explicit code points so the test does not depend
	// on how this source file happens to be encoded on disk: precomposed "\u00e9"
	// (é) vs "e" + combining acute accent "\u0301".
	composed := "Caf\u00e9"
	decomposed := "Cafe\u0301"
	if composed == decomposed {
		t.Fatal("test setup wrong: the two forms are already byte-equal")
	}
	if NormalizeName(composed) != NormalizeName(decomposed) {
		t.Errorf("NFC: %q and %q normalize differently (%q vs %q)",
			composed, decomposed, NormalizeName(composed), NormalizeName(decomposed))
	}
}
