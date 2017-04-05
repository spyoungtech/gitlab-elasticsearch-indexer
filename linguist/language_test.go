package linguist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/es-git-go/linguist"
)

func TestCommonLanguagesAreDetectedByExtension(t *testing.T) {
	type tc struct {
		filename string
		name     string
	}

	for _, tc := range []struct {
		file string
		lang string
	}{
		{"foo.go", "Go"},
		{".go", "Go"},
		{"foo.go.rb", "Ruby"},
		{"foo.rb", "Ruby"},
		{"foo.c", "C"},
		{"foo.cpp", "C++"},
		{"/bar/foo.ini", "INI"},
		{"bar/foo.ini", "INI"},
		{"c:/foo.ini", "INI"},
		{`c:\foo.ini`, "INI"},
		{"foo.md", "Markdown"}, // Multiple possible languages
	} {
		langs := linguist.DetectLanguageByExtension(tc.file)
		assert.Equal(t, 1, len(langs))
		assert.Equal(t, tc.lang, langs[0].Name)

		lang := linguist.DetectLanguage(tc.file, []byte{})
		assert.NotNil(t, lang)
		assert.Equal(t, tc.lang, lang.Name)

	}
}

func TestImaginaryLanguageIsntRecognised(t *testing.T) {
	lang := linguist.DetectLanguageByFilename("foo.absolutely-nobody-will-make-this-extension")
	assert.Nil(t, lang)
}

// This test checks the content of languages.go against expectations chosen to
// validate the go:generate script
func TestAttributesAreCopiedCorrectly(t *testing.T) {
	ada := linguist.Languages["Ada"]
	assert.NotNil(t, ada)

	cmake := linguist.Languages["CMake"]
	assert.NotNil(t, cmake)

	gettext := linguist.Languages["Gettext Catalog"]
	assert.NotNil(t, gettext)

	golang := linguist.Languages["Go"]
	assert.NotNil(t, golang)

	json := linguist.Languages["JSON"]
	assert.NotNil(t, json)

	markdown := linguist.Languages["Markdown"]
	assert.NotNil(t, markdown)

	ruby := linguist.Languages["Ruby"]
	assert.NotNil(t, ruby)

	assert.Equal(t, "Go", golang.Name)
	assert.Equal(t, "programming", golang.Type)
	assert.Equal(t, "JavaScript", json.Group)
	assert.Equal(t, "#375eab", golang.Color)
	assert.Equal(t, []string{"ada95", "ada2005"}, ada.Aliases)
	assert.Equal(t, []string{".go"}, golang.Extensions)
	assert.Equal(t, []string{"CMakeLists.txt"}, cmake.Filenames)
	assert.Equal(t, []string{"ruby", "macruby", "rake", "jruby", "rbx"}, ruby.Interpreters)
	assert.Equal(t, "source.gfm", markdown.TmScope)
	assert.Equal(t, "programming", golang.AceMode)
	assert.Equal(t, false, gettext.Searchable)
	assert.Equal(t, true, markdown.Wrap)

}
