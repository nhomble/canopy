package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const maxLinesPerFile = 10000

type ImportStatement struct {
	Raw    string
	Source string
}

type Signature struct {
	Kind string
	Name string
	Line int
	Raw  string
}

type FileSignatures struct {
	Path       string
	RelPath    string
	Language   string
	Imports    []ImportStatement
	Signatures []Signature
	Decorators []string
}

func ExtractSignatures(filePath, relPath string) (*FileSignatures, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	lang := GetLanguage(ext)
	if lang == nil {
		return nil, nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	result := &FileSignatures{
		Path:     filePath,
		RelPath:  relPath,
		Language: lang.Name,
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	seenDecorators := make(map[string]bool)
	for sc.Scan() {
		lineNum++
		if lineNum > maxLinesPerFile {
			break
		}
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		for _, importRe := range lang.Imports {
			matches := importRe.FindStringSubmatch(trimmed)
			if matches != nil {
				source := ""
				for i := len(matches) - 1; i >= 1; i-- {
					if matches[i] != "" {
						source = matches[i]
						break
					}
				}
				if source != "" {
					result.Imports = append(result.Imports, ImportStatement{Raw: trimmed, Source: source})
				}
				break
			}
		}
		for _, sigPattern := range lang.Signatures {
			matches := sigPattern.Pattern.FindStringSubmatch(trimmed)
			if matches != nil && len(matches) > 1 {
				result.Signatures = append(result.Signatures, Signature{
					Kind: sigPattern.Kind, Name: matches[1], Line: lineNum, Raw: trimmed,
				})
				break
			}
		}
		if lang.Decorators != nil {
			matches := lang.Decorators.FindAllStringSubmatch(trimmed, -1)
			for _, m := range matches {
				if len(m) > 1 && !seenDecorators[m[1]] {
					seenDecorators[m[1]] = true
					result.Decorators = append(result.Decorators, m[1])
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
