package zfmt

import (
	"log"
	"path/filepath"

	"golang.org/x/tools/imports"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
)

var (
	jsFormatOptions = jsbeautifier.DefaultOptions()
	goFormatOptions = &imports.Options{
		TabWidth:  4,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	}
)

// absPath returns the absolute path of this file
func absPath(fPath string) string {
	fPathAbsolute, err := filepath.Abs(fPath)
	if err != nil {
		log.Fatal(err)
	}
	return fPathAbsolute
}

func FormatCSS(fPath string, fContent []byte) []byte {
	resString := formatCSS(string(fContent))
	return []byte(resString)
}

func FormatJS(fPath string) []byte {
	resString := jsbeautifier.BeautifyFile(fPath, jsFormatOptions)
	return []byte(*resString)
}

func FormatGo(fPath string, fContent []byte) []byte {
	fPathAbsolute := absPath(fPath)
	result, err := imports.Process(fPathAbsolute, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	return result
}
