package thaiaddress

// Province is a Thai province (จังหวัด). Bangkok (code 10) is administratively a
// special area but is modelled as a province here.
type Province struct {
	Code   int    `json:"code"`   // official 2-digit DOPA code, 10–96
	NameTH string `json:"nameTh"` // e.g. "กรุงเทพมหานคร"
	NameEN string `json:"nameEn"` // e.g. "Bangkok"
	Region Region `json:"region"` // marshals as the integer Region code
}

// District is a Thai district: อำเภอ in most provinces, เขต in Bangkok.
type District struct {
	Code         int    `json:"code"`         // official 4-digit DOPA code; Code/100 == ProvinceCode
	ProvinceCode int    `json:"provinceCode"` // owning province's 2-digit code
	NameTH       string `json:"nameTh"`       // includes the "อำเภอ"/"เขต" sense via the source naming
	NameEN       string `json:"nameEn"`
}

// Subdistrict is a Thai subdistrict: ตำบล in most provinces, แขวง in Bangkok.
type Subdistrict struct {
	Code         int    `json:"code"`         // official 6-digit DOPA code; Code/100 == DistrictCode
	DistrictCode int    `json:"districtCode"` // owning district's 4-digit code
	NameTH       string `json:"nameTh"`
	NameEN       string `json:"nameEn"`
	Postcode     int    `json:"postcode"` // 5-digit Thai postal code (รหัสไปรษณีย์)
}

// Districts returns the districts of p, ordered by code. Returns nil for an
// unknown province.
func (p Province) Districts() []District { return DistrictsOf(p.Code) }

// Province returns the province a district belongs to.
func (d District) Province() (Province, bool) { return ProvinceByCode(d.ProvinceCode) }

// Subdistricts returns the subdistricts of d, ordered by code. Returns nil for
// an unknown district.
func (d District) Subdistricts() []Subdistrict { return SubdistrictsOf(d.Code) }

// Postcodes returns the distinct postal codes used within d, ordered ascending.
func (d District) Postcodes() []int { return PostcodesOf(d.Code) }

// District returns the district a subdistrict belongs to.
func (s Subdistrict) District() (District, bool) { return DistrictByCode(s.DistrictCode) }

// Province returns the province a subdistrict belongs to.
func (s Subdistrict) Province() (Province, bool) { return ProvinceByCode(s.DistrictCode / 100) }
