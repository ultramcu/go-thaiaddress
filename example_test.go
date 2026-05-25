package thaiaddress_test

import (
	"errors"
	"fmt"

	thaiaddress "github.com/ultramcu/go-thaiaddress"
)

func ExampleProvinceByCode() {
	p, _ := thaiaddress.ProvinceByCode(10)
	fmt.Printf("%s / %s — %s\n", p.NameEN, p.NameTH, p.Region)
	// Output: Bangkok / กรุงเทพมหานคร — Central
}

func ExampleProvince_Districts() {
	cm, _ := thaiaddress.ProvinceByCode(50) // Chiang Mai
	fmt.Println(len(cm.Districts()))
	// Output: 25
}

func ExampleByPostcode() {
	// A Thai postal code commonly covers several subdistricts.
	for _, s := range thaiaddress.ByPostcode(50200) {
		p, _ := s.Province()
		fmt.Printf("%s, %s\n", s.NameEN, p.NameEN)
	}
	// Output:
	// Si Phum, Chiang Mai
	// Phra Sing, Chiang Mai
	// Suthep, Chiang Mai
}

func ExampleResolve() {
	matches, err := thaiaddress.Resolve(thaiaddress.Query{
		Subdistrict: "ในเมือง",
		District:    "เมืองขอนแก่น",
		Province:    "ขอนแก่น",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	m := matches[0]
	fmt.Printf("%d %s %d\n", m.Subdistrict.Code, m.Subdistrict.NameEN, m.Subdistrict.Postcode)
	// Output: 400101 Nai Mueang 40000
}

func ExampleValidate() {
	fmt.Println(thaiaddress.Validate(10, 1001, 100101)) // Bangkok / Phra Nakhon / ...
	err := thaiaddress.Validate(10, 5001, 100101)       // district 5001 is Chiang Mai
	fmt.Println(errors.Is(err, thaiaddress.ErrInconsistent))
	// Output:
	// <nil>
	// true
}

func ExampleNormalizeName() {
	// Leading administrative prefixes (Thai or English) are stripped so user
	// input matches the stored names.
	fmt.Println(thaiaddress.NormalizeName("อำเภอเมืองเชียงใหม่"))
	fmt.Println(thaiaddress.NormalizeName("เขตพระนคร"))
	// Output:
	// เมืองเชียงใหม่
	// พระนคร
}

func ExampleFindSubdistricts() {
	// Subdistrict names repeat across provinces, so Find* returns every match.
	subs := thaiaddress.FindSubdistricts("ในเมือง")
	fmt.Println(len(subs) > 1)
	// Output: true
}
