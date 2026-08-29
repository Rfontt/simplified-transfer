package user

import "strings"

type Document string

func NewDocument(value string) (Document, error) {
	digits := onlyDigits(value)
	switch len(digits) {
	case 11:
		if validCPF(digits) {
			return Document(digits), nil
		}
	case 14:
		if validCNPJ(digits) {
			return Document(digits), nil
		}
	}
	return "", &InvalidDocumentError{Document: value}
}

func (d Document) String() string {
	return string(d)
}

var (
	cpfWeights1  = []int{10, 9, 8, 7, 6, 5, 4, 3, 2}
	cpfWeights2  = []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	cnpjWeights1 = []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	cnpjWeights2 = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
)

func validCPF(digits string) bool {
	if allSame(digits) {
		return false
	}
	return digits[9] == checkDigit(digits[:9], cpfWeights1) &&
		digits[10] == checkDigit(digits[:10], cpfWeights2)
}

func validCNPJ(digits string) bool {
	if allSame(digits) {
		return false
	}
	return digits[12] == checkDigit(digits[:12], cnpjWeights1) &&
		digits[13] == checkDigit(digits[:13], cnpjWeights2)
}

func checkDigit(partial string, weights []int) byte {
	sum := 0
	for i := 0; i < len(partial); i++ {
		sum += int(partial[i]-'0') * weights[i]
	}
	rem := sum % 11
	if rem < 2 {
		return '0'
	}
	return byte('0' + 11 - rem)
}

func allSame(digits string) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			return false
		}
	}
	return true
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
