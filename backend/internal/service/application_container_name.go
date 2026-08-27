package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
)

const containerNameTimestampLayout = "2006-01-02-15-04-05"

func buildApplicationContainerName(applicantName string, createdAt time.Time) (string, error) {
	namePinyin := applicantNamePinyin(applicantName)
	if namePinyin == "" {
		return "", errors.New("applicant name cannot be converted to a valid container name")
	}
	return fmt.Sprintf("%s-%s", namePinyin, createdAt.Format(containerNameTimestampLayout)), nil
}

func applicantNamePinyin(applicantName string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	args.Heteronym = false
	args.Fallback = func(character rune, _ pinyin.Args) []string {
		switch {
		case character >= 'A' && character <= 'Z':
			return []string{string(character + ('a' - 'A'))}
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			return []string{string(character)}
		default:
			return nil
		}
	}

	var name strings.Builder
	for _, syllable := range pinyin.LazyPinyin(strings.TrimSpace(applicantName), args) {
		for _, character := range strings.ToLower(syllable) {
			if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
				name.WriteRune(character)
			}
		}
	}
	return name.String()
}
