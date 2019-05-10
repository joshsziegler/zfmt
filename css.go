package zfmt

import (
	"regexp"
	"strings"
)

func cssRemoveVariables(css string) string {
	// Find all of the CSS variable declarations
	cssVarDefinitions := regexp.MustCompile("--(?P<var_name>[\\w-]+?):\\s+?(?P<var_val>.+?);")
	match := cssVarDefinitions.FindAllStringSubmatch(css, -1)
	// Some CSS variables use others nested CSS variables in their definition. Replace each by working backwards.
	for i := len(match) - 1; i >= 0; i-- {
		re := regexp.MustCompile("var\\(--" + match[i][1] + "\\)")
		css = re.ReplaceAllString(css, match[i][2])
	}
	// Remove the :root{} section with all of the CSS variable declarations
	re := regexp.MustCompile(":root \\{[\\s\\S]*?\\}")
	css = re.ReplaceAllString(css, "")

	return css
}

func cssRemoveNestedCalc(css string) string {
	// Find nested calc() expressions and make the inner expression a simple parenthesis
	// Example:
	//     calc(100% - calc(2rem + 2px));
	//     The above should become:
	//     calc(100% - (2rem +2px));

	// We want to keep the text inbetween the two 'calc(' parts
	reNestedCalc := regexp.MustCompile("calc\\((.*?)calc\\(")
	matches := reNestedCalc.FindAllStringSubmatchIndex(css, -1)
	// Work backward so that after we modify the CSS, the match indices still match correctly
	for i := len(matches) - 1; i >= 0; i-- {
		v := matches[i]
		css = css[0:v[0]] + "calc(" + css[v[2]:v[3]] + "(" + css[v[1]:len(css)]
	}
	return css

}

func cssMinimize(css string) string {
	// Remove Comments
	commentReg := regexp.MustCompile(`\s*\/\*[\s\S]*?\*\/\s*`)
	css = commentReg.ReplaceAllString(css, "")

	//Protect Strings
	stringReg := regexp.MustCompile(`(content\s*:|[\w-]+\s*=)\s*(([\'\"]).*?([\'\"]))\s*`)
	stringSlice := stringReg.FindAllString(css, -1)
	css = stringReg.ReplaceAllString(css, `${1}!string!`)

	// Protect urls
	urlReg := regexp.MustCompile(`((?:url|url-prefix|regexp)\([^\)]+\))`)
	urls := urlReg.FindAllString(css, -1)
	css = urlReg.ReplaceAllString(css, `!url`)

	// Pre process
	re := regexp.MustCompile(`\s*([\{\}:;,])\s*`) // Remove \s before and after characters {}:;,
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`([\[\(])\s*`) // Remove space inner [ or (
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`\s*([\)\]])`) // Remove space inner ) or ]
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`,[\d\s\.\#\+>~:]*\{`) // Remove invalid selectors without \w
	css = re.ReplaceAllString(css, "{")
	re = regexp.MustCompile(`([;,])([;,])+`) // Remove repeated ;,
	css = re.ReplaceAllString(css, "${1}")

	// Process action rules
	css = cssCompressRules(css)

	css = strings.Replace(css, ";}", "}", -1)

	// Backfill strings
	for _, str := range stringSlice {
		css = replaceFirstInstance(css, "!string!", string(str[1]))
	}

	// Backfill urls
	for _, url := range urls {
		css = replaceFirstInstance(css, "!url", url)
	}

	// Trim
	re = regexp.MustCompile(`^\s*(\S+(\s+\S+)*)\s*$`)
	css = re.ReplaceAllString(css, "${1}")

	return css
}

func cssCompressRules(css string) string {
	re := regexp.MustCompile(`\s*([\{\}:;,])\s*`)
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`\s+!important`)
	css = re.ReplaceAllString(css, "!important")
	re = regexp.MustCompile(`((?:@charset|@import)[^;]+;)\s*`)
	css = re.ReplaceAllString(css, "${1]\n")

	return css
}

func formatCSS(css string) string {
	// Protect Comments
	commentReg := regexp.MustCompile(`[ \t]*\/\*[\s\S]*?\*\/`)
	comments := commentReg.FindAllString(css, -1)
	css = commentReg.ReplaceAllString(css, "!comment!")

	//Protect Strings
	stringReg := regexp.MustCompile(`(content\s*:|[\w-]+\s*=)\s*(([\'\"]).*?([\'\"]))\s*`)
	stringSlice := stringReg.FindAllString(css, -1)
	css = stringReg.ReplaceAllString(css, "${1}!string!")

	// Protect urls
	urlReg := regexp.MustCompile(`((?:url|url-prefix|regexp)\([^\)]+\))`)
	urls := urlReg.FindAllString(css, -1)
	css = urlReg.ReplaceAllString(css, `!url`)

	// Pre process
	re := regexp.MustCompile(`\s*([\{\}:;,])\s*`) // Remove \s before and after characters {}:;,
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`([\[\(])\s*`) // Remove space inner [ or (
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`\s*([\)\]])`) // Remove space inner ) or ]
	css = re.ReplaceAllString(css, "${1}")
	re = regexp.MustCompile(`,[\d\s\.\#\+>~:]*\{`) // Remove invalid selectors without \w
	css = re.ReplaceAllString(css, "{")
	re = regexp.MustCompile(`([;,])([;,])+`) // Remove repeated ;,
	css = re.ReplaceAllString(css, "${1}")

	// Group selector
	css = cssBreakSelectors(css) // Break after selectors' ,

	// Add space
	re = regexp.MustCompile(`([A-Za-z-](?:\+_?)?):([^;\{]+[;\}])`) // Add space after properties' :
	css = re.ReplaceAllString(css, "${1}: ${2}")

	// Process action rules
	css = cssExpandRules(css)

	// Add blank line between each block
	re = regexp.MustCompile(`\}\s*`)
	css = re.ReplaceAllString(css, "}\n\n")

	// Fix comments
	re = regexp.MustCompile(`\s*!comment!\s*@`)
	css = re.ReplaceAllString(css, "\n\n!comment!\n@")
	re = regexp.MustCompile(`\s*!comment!\s*([^\/\{\};]+?)\{`)
	css = re.ReplaceAllString(css, "\n\n!comment!\n${1}{")
	re = regexp.MustCompile(`\s*\n!comment!`)
	css = re.ReplaceAllString(css, "\n\n!comment!")

	// Backfill comments
	for _, comment := range comments {
		css = replaceFirstInstance(css, "[ \t]*!comment!", comment)
	}

	// Indent
	css = indentCode(css, "    ")

	// Backfill strings
	for _, str := range stringSlice {
		re = regexp.MustCompile(`(content\s*:|[\w-]+\s*=)\s*(([\'\"]).*?([\'\"]))\s*`)
		temp := re.ReplaceAllString(str, "${2}")
		css = replaceFirstInstance(css, "!string!", temp)
	}

	// Backfill urls
	for _, url := range urls {
		css = replaceFirstInstance(css, "!url", url)
	}

	// Trim
	re = regexp.MustCompile(`^\s*(\S+(\s+\S+)*)\s*$`)
	css = re.ReplaceAllString(css, "${1}")

	return css
}

func replaceFirstInstance(css string, reg string, repl string) string {
	re := regexp.MustCompile(reg)
	i := 1
	css = re.ReplaceAllStringFunc(css, func(s string) string {
		if i != 0 {
			i -= 1
			return repl
		}
		return s
	})
	return css
}

func cssExpandRules(css string) string {
	re := regexp.MustCompile(`{`)
	css = re.ReplaceAllString(css, " {\n")

	re = regexp.MustCompile(`;`)
	css = re.ReplaceAllString(css, ";\n")
	re = regexp.MustCompile(`;\s*([^\{\};]+?)\{`)
	css = re.ReplaceAllString(css, ";\n\n${1}{")

	re = regexp.MustCompile(`\s*(!comment!)\s*;\s*`)
	css = re.ReplaceAllString(css, " !comment! ;\n")
	re = regexp.MustCompile(`(:[^:;]+;)\s*(!comment!)\s*`)
	css = re.ReplaceAllString(css, "${1} ${2}\n")

	re = regexp.MustCompile(`\s*\}`)
	css = re.ReplaceAllString(css, "\n}")
	re = regexp.MustCompile(`\}\s*`)
	css = re.ReplaceAllString(css, "}\n")

	return css
}

func cssBreakSelectors(css string) string {
	block := strings.Split(css, "}")
	for i := 0; i < len(block); i++ {
		b := strings.Split(block[i], "{")
		bLen := len(b)

		for j := 0; j < bLen; j++ {
			if j == bLen-1 {
				re := regexp.MustCompile(`,\s*`)
				b[j] = re.ReplaceAllString(b[j], ", ")
			} else {
				s := strings.Split(b[j], ";")
				sLen := len(s)
				sLast := s[sLen-1]

				for k := 0; k < sLen-1; k++ {
					re := regexp.MustCompile(`,\s*`)
					s[k] = re.ReplaceAllString(s[k], ", ")
				}

				re1 := regexp.MustCompile(`\s*@(document|media)`)
				re2 := regexp.MustCompile(`(\(|\))`)
				if re1.FindString(sLast) != "" {
					re := regexp.MustCompile(`,\s*`)
					s[sLen-1] = re.ReplaceAllString(sLast, ", ")
				} else if re2.FindString(sLast) != "" {
					u := strings.Split(sLast, ")")
					for m := 0; m < len(u); m++ {
						v := strings.Split(u[m], "(")
						vLen := len(v)
						if vLen < 2 {
							continue
						}
						re := regexp.MustCompile(`,\s*`)
						v[0] = re.ReplaceAllString(v[0], ",\n")
						v[1] = re.ReplaceAllString(v[1], ", ")
						u[m] = strings.Join(v, "(")
					}
					s[sLen-1] = strings.Join(u, ")")
				} else {
					re := regexp.MustCompile(`,\s*`)
					s[sLen-1] = re.ReplaceAllString(sLast, ",\n")
				}
				b[j] = strings.Join(s, ";")
			}
		}
		block[i] = strings.Join(b, "{")
	}
	css = strings.Join(block, "}")

	return css
}

func indentCode(css string, indentation string) string {
	lines := strings.Split(css, "\n")
	level := 0
	inComment := false
	outPrefix := ""
	adjustment := 0

	for i := 0; i < len(lines); i++ {
		if !inComment {
			// Quote level adjustment
			re := regexp.MustCompile(`\/\*[\s\S]*?\*\/`)
			validCode := re.ReplaceAllString(lines[i], "")
			re = regexp.MustCompile(`\/\*[\s\S]*`)
			validCode = re.ReplaceAllString(validCode, "")
			adjustment = strings.Count(validCode, "{") - strings.Count(validCode, "}")

			// Trim
			re = regexp.MustCompile(`^(\s+)\/\*.*`)
			m := re.FindStringSubmatch(lines[i])
			if m != nil {
				outPrefix := m[1]
				re := regexp.MustCompile(`^` + outPrefix + `(.*)\s*$`)
				lines[i] = re.ReplaceAllString(lines[i], "${1}")
			} else {
				re := regexp.MustCompile(`^\s*(.*)\s*$`)
				lines[i] = re.ReplaceAllString(lines[i], "${1}")
			}
		} else {
			// Quote level adjustment
			adjustment = 0

			// Trim
			re := regexp.MustCompile(`^` + outPrefix + `(.*)\s*$`)
			lines[i] = re.ReplaceAllString(lines[i], "${1}")
		}
		// Is next line in comment
		re := regexp.MustCompile(`\/\*|\*\/`)
		commentQuotes := re.FindAllString(lines[i], -1)
		for _, quote := range commentQuotes {
			if inComment && quote == "*/" {
				inComment = false
			} else if quote == "/*" {
				inComment = true
			}
		}

		// Quote level adjustment
		nextLevel := level + adjustment
		thisLevel := 0
		if adjustment > 0 {
			thisLevel = level
		} else {
			thisLevel = nextLevel
		}
		level = nextLevel

		// Add indentation
		if lines[i] != "" {
			lines[i] = strings.Repeat(indentation, thisLevel) + lines[i]
		} else {
			lines[i] = ""
		}
	}
	css = strings.Join(lines, "\n")

	return css
}
