package indexer

import (
	"fmt"
	"log"

	"github.com/lupine/icu"
)

var (
	detector  *icu.CharsetDetector
	converter = icu.NewCharsetConverter(maxBlobSize)
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

	return "", fmt.Errorf("Failed to convert from %s to UTF-8", matches[0].Charset)
}
