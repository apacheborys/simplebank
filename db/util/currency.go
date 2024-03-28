package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}

func PickOtherCurrency(currency string) string {
	switch currency {
	case USD:
		return EUR
	case EUR:
		return CAD
	case CAD:
		return USD
	}
	return ""
}
