package url

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const regExBlockList = "(?:facebook)"

// checkBlockList custom rule for block list validation
func checkBlockList(value interface{}) error {
	s, _ := value.(string)
	match, _ := regexp.MatchString(regExBlockList, s)
	if match {
		return errors.New("url is not allowed")
	}
	return nil
}

// generateRandomString: Helper function to create random string for shorten service
func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
