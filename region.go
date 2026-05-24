package thaiaddress

// Region is one of Thailand's six conventional geographic regions, as used by
// the Royal Society's classification.
type Region int

const (
	RegionNorth     Region = 1 // ภาคเหนือ
	RegionCentral   Region = 2 // ภาคกลาง
	RegionNortheast Region = 3 // ภาคตะวันออกเฉียงเหนือ (อีสาน)
	RegionWest      Region = 4 // ภาคตะวันตก
	RegionEast      Region = 5 // ภาคตะวันออก
	RegionSouth     Region = 6 // ภาคใต้
)

var regionNames = map[Region][2]string{
	RegionNorth:     {"ภาคเหนือ", "North"},
	RegionCentral:   {"ภาคกลาง", "Central"},
	RegionNortheast: {"ภาคตะวันออกเฉียงเหนือ", "Northeast"},
	RegionWest:      {"ภาคตะวันตก", "West"},
	RegionEast:      {"ภาคตะวันออก", "East"},
	RegionSouth:     {"ภาคใต้", "South"},
}

// Valid reports whether r is one of the six known regions.
func (r Region) Valid() bool { _, ok := regionNames[r]; return ok }

// NameTH returns the Thai region name, or "" if r is unknown.
func (r Region) NameTH() string { return regionNames[r][0] }

// NameEN returns the English region name, or "" if r is unknown.
func (r Region) NameEN() string { return regionNames[r][1] }

// String returns the English region name (implements fmt.Stringer).
func (r Region) String() string {
	if n := r.NameEN(); n != "" {
		return n
	}
	return "Region(unknown)"
}
