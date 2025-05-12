package utils

import (
	"regexp"
)

// IsValidVietnamesePhone checks if a string is a valid Vietnamese phone number
// Valid formats:
// - Mobile: 10 digits starting with 03, 05, 07, 08, 09 (e.g., 0912345678)
// - Landline: 10-11 digits including area code (e.g., 02812345678)
func IsValidVietnamesePhone(phone string) bool {
	if phone == "" {
		return false
	}

	// Mobile phone patterns (post-2018)
	mobilePattern := `^(03|05|07|08|09)[0-9]{8}$`

	// Some older mobile formats (pre-2018 Viettel, Mobifone, Vinaphone)
	oldMobilePattern := `^(01[2689])[0-9]{8}$`

	// Landline patterns (area codes + local number)
	landlinePattern := `^(02[0-9])[0-9]{7,8}$` // For most provinces/cities

	// Combined pattern
	fullPattern := regexp.MustCompile(mobilePattern + "|" + oldMobilePattern + "|" + landlinePattern)

	return fullPattern.MatchString(phone)
}
