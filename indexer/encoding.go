package indexer

import (
	"fmt"
	"log"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/git"

	"gitlab.com/lupine/icu"
)

var (
	detector  *icu.CharsetDetector
	converter = icu.NewCharsetConverter(git.LimitFileSize)
)

func init() {
	var err error
	detector, err = icu.NewCharsetDetector()
	if err != nil {
		panic(err)
	}
}

func tryEncodeString(s string) string {
	encoded, err := encodeString(s)
	if err != nil {
		log.Println(err)
		return s // TODO: Run it through the UTF-8 replacement encoder
	}

	return encoded
}

func tryEncodeBytes(b []byte) string {
	encoded, err := encodeBytes(b)
	if err != nil {
		log.Println(err)
		s := string(b)
		return s // TODO: Run it through the UTF-8 replacement encoder
	}

	return encoded
}

func encodeString(s string) (string, error) {
	return encodeBytes([]byte(s))
}

// encodeString converts a string from an arbitrary encoding to UTF-8
func encodeBytes(b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}

	matches, err := detector.GuessCharset(b)
	if err != nil {
		return "", fmt.Errorf("Couldn't guess charset: %s", err)
	}

	// Try encoding for each match, returning the first that succeeds
	for _, match := range matches {
		utf8, err := converter.ConvertToUtf8(b, match.Charset)
		if err == nil {
			return string(utf8), nil
		}
	}

	// `detector.GuessCharset` may return err == nil && len(matches) == 0
	bestGuess := "unknown"
	if len(matches) > 0 {
		bestGuess = matches[0].Charset
	}

	return "", fmt.Errorf("Failed to convert from %s to UTF-8", bestGuess)
}
