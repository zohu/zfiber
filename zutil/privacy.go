package zutil

import "fmt"

func PrivacyName(name string) string {
	arr := []rune(name)
	if len(arr) <= 2 {
		return fmt.Sprintf("%s**", string(arr[0]))
	}
	return fmt.Sprintf("%s*%s", string(arr[0]), string(arr[len(arr)-1]))
}
func PrivacyPhone(phone string) string {
	arr := []rune(phone)
	return fmt.Sprintf("%s****%s", string(arr[0]), string(arr[len(arr)-4:]))
}
