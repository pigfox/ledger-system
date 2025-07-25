package config

import "regexp"

func IsValidCurrency(currency string) bool {
	return AllowedCurrencies[currency]
}

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
