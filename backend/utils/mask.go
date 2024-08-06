package utils

import "strings"

var mask = "*"

func MaskEmail(email string) string {
	if len(email) == 0 {
		return ""
	}

	tmp := strings.Split(email, "@")

	name := tmp[0]
	domain := tmp[1]

	nameRunes := []rune(name)
	nameRunesLen := len(nameRunes)

	if nameRunesLen == 0 {
		return strings.Repeat(mask, 6) + "@" + domain
	}

	var maskStart, padLength int
	if nameRunesLen <= 6 {
		maskStart = 1
		padLength = 6 - nameRunesLen
	} else {
		maskStart = 3

	}

	maskedAddress := ""
	maskedAddress += string(nameRunes[:maskStart])
	maskedAddress += strings.Repeat(mask, len(nameRunes[maskStart:])+padLength)
	maskedAddress += "@" + domain

	return maskedAddress
}

func MaskUsername(username string) string {
	usernameRunes := []rune(username)
	usernameRunesLen := len(usernameRunes)

	if usernameRunesLen == 0 {
		return ""
	}

	if usernameRunesLen == 1 {
		return mask
	}

	padLength := 0
	if usernameRunesLen <= 3 {
		padLength = 6 - usernameRunesLen
	}

	maskedUsername := ""
	maskedUsername += string(usernameRunes[0])
	maskedUsername += strings.Repeat(mask, usernameRunesLen-2+padLength)
	maskedUsername += string(usernameRunes[usernameRunesLen-1])

	return maskedUsername
}
