package linguist

import (
	"path"
)

// There's no YAML support in the Go stdlib, so use ruby instead.
// This will create a file `languages.go`, containing the Languages variable
//go:generate ruby generate_languages.go.rb

type Language struct {
	Name         string
	Type         string
	Group        string
	Color        string
	Aliases      []string
	Extensions   []string
	Filenames    []string
	Interpreters []string
	TmScope      string
	AceMode      string
	Wrap         bool
	Searchable   bool
}

var (
	languagesByExtension map[string][]*Language
	languagesByFilename  map[string][]*Language
)

func init() {
	languagesByExtension = make(map[string][]*Language)
	for _, lang := range Languages {
		for _, ext := range lang.Extensions {
			languagesByExtension[ext] = append(languagesByExtension[ext], lang)
		}
	}

	languagesByFilename = make(map[string][]*Language)
	for _, lang := range Languages {
		for _, filename := range lang.Filenames {
			languagesByFilename[filename] = append(languagesByFilename[filename], lang)
		}
	}
}

// and returns only the languges present in both A and B
func and(a, b []*Language) []*Language {
	var out []*Language

	for _, langA := range a {
		for _, langB := range b {
			if langA == langB {
				out = append(out, langA)
			}
		}
	}

	return out
}

func DetectLanguageByFilename(filename string) []*Language {
	return languagesByFilename[path.Base(filename)]
}

func DetectLanguageByExtension(filename string) []*Language {
	return languagesByExtension[path.Ext(filename)]
}

func DetectLanguage(filename string, blob []byte) *Language {
	// TODO: github-linguist uses a range of strategies not replicated here.
	// It does the following:
	//
	//   * modelines
	//   * shebangs
	//   * filename / extension (we have these)
	//   * heuristics
	//   * classifier

	byFilename := DetectLanguageByFilename(filename)
	if len(byFilename) == 1 {
		return byFilename[0]
	}

	byExtension := DetectLanguageByExtension(filename)
	if len(byFilename) > 1 {
		byExtension = and(byFilename, byExtension)
	}

	if len(byExtension) > 0 {
		return byExtension[0]
	}

	return nil
}
