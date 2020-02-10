package linguist_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-elasticsearch-indexer/linguist"
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
		require.Equal(t, 1, len(langs))
		require.Equal(t, tc.lang, langs[0].Name)

		lang := linguist.DetectLanguage(tc.file, []byte{})
		require.NotNil(t, lang)
		require.Equal(t, tc.lang, lang.Name)

	}
}

func TestImaginaryLanguageIsntRecognised(t *testing.T) {
	lang := linguist.DetectLanguageByFilename("foo.absolutely-nobody-will-make-this-extension")
	require.Nil(t, lang)
}

// This test checks the content of languages.go against expectations chosen to
// validate the go:generate script
func TestAttributesAreCopiedCorrectly(t *testing.T) {
	ada := linguist.Languages["Ada"]
	require.NotNil(t, ada)

	cmake := linguist.Languages["CMake"]
	require.NotNil(t, cmake)

	gettext := linguist.Languages["Gettext Catalog"]
	require.NotNil(t, gettext)

	golang := linguist.Languages["Go"]
	require.NotNil(t, golang)

	json := linguist.Languages["JSON"]
	require.NotNil(t, json)

	markdown := linguist.Languages["Markdown"]
	require.NotNil(t, markdown)

	ruby := linguist.Languages["Ruby"]
	require.NotNil(t, ruby)

	require.Equal(t, "Go", golang.Name)
	require.Equal(t, "programming", golang.Type)
	require.Equal(t, "JavaScript", json.Group)
	require.Equal(t, "#375eab", golang.Color)
	require.Equal(t, []string{"ada95", "ada2005"}, ada.Aliases)
	require.Equal(t, []string{".go"}, golang.Extensions)
	require.Equal(t, []string{"CMakeLists.txt"}, cmake.Filenames)
	require.Equal(t, []string{"ruby", "macruby", "rake", "jruby", "rbx"}, ruby.Interpreters)
	require.Equal(t, "source.gfm", markdown.TmScope)
	require.Equal(t, "programming", golang.AceMode)
	require.Equal(t, false, gettext.Searchable)
	require.Equal(t, true, markdown.Wrap)

}
