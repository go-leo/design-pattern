package internal

import (
	"log"
	"path"
	"regexp"
	"strings"
	"unicode"
)

type annotation string

const (
	CQRS        annotation = "@CQRS"
	Query       annotation = "@Query"
	Command     annotation = "@Command"
	QueryPath   annotation = "@QueryPath"
	CommandPath annotation = "@CommandPath"
)

func (a annotation) String() string {
	return string(a)
}

func (a annotation) EqualsIgnoreCase(str string) bool {
	return strings.ToUpper(str) == strings.ToUpper(a.String())
}

func (a annotation) PrefixOf(str string) bool {
	return strings.HasPrefix(strings.ToUpper(str), strings.ToUpper(a.String()))
}

type Path struct {
	Query   string
	Command string
}

func NewPath(comments []string) *Path {
	info := &Path{}
	for _, comment := range comments {
		text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "//"))
		seg := strings.Split(text, " ")
		// 注释的开始必须以 @CQRS 开头
		if !CQRS.EqualsIgnoreCase(seg[0]) {
			continue
		}
		for _, s := range seg {
			s = strings.TrimSpace(s)
			switch {
			case QueryPath.PrefixOf(s):
				v, ok := ExtractValue(s, string(QueryPath))
				if !ok {
					log.Fatalf("error: %s query path invalid", s)
				}
				info.Query = v
			case CommandPath.PrefixOf(s):
				v, ok := ExtractValue(s, string(CommandPath))
				if !ok {
					log.Fatalf("error: %s command path invalid", s)
				}
				info.Command = v
			}
		}
	}
	return info
}

func NewFileFromComment(endpoint string, queryDir, commandDir string, comments []string) *File {
	for _, comment := range comments {
		text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "//"))
		seg := strings.Split(text, " ")
		// 注释的开始必须以 @CQRS 开头
		if !CQRS.EqualsIgnoreCase(seg[0]) {
			continue
		}
		for _, s := range seg {
			s = strings.TrimSpace(s)
			switch {
			case Query.EqualsIgnoreCase(s):
				return NewQueryFile(endpoint, queryDir)
			case Command.EqualsIgnoreCase(s):
				return NewCommandFile(endpoint, commandDir)
			}
		}
	}
	return nil
}

func NewQueryFile(endpoint string, queryDir string) *File {
	r := &File{
		Type:          "query",
		AbsFilename:   path.Join(queryDir, strings.ToLower(addUnderscore(endpoint))+".go"),
		Package:       path.Base(queryDir),
		Endpoint:      endpoint,
		LowerEndpoint: strings.ToLower(endpoint[:1]) + endpoint[1:],
	}
	return r
}

func NewCommandFile(endpoint string, commandDir string) *File {
	r := &File{
		Type:          "command",
		AbsFilename:   path.Join(commandDir, strings.ToLower(addUnderscore(endpoint))+".go"),
		Package:       path.Base(commandDir),
		Endpoint:      endpoint,
		LowerEndpoint: strings.ToLower(endpoint[:1]) + endpoint[1:],
	}
	return r
}

func addUnderscore(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(rune(s[i-1])) {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return result.String()
}

func ExtractValue(s string, annotation string) (string, bool) {
	reg := regexp.MustCompile(annotation + `\((.*)\)`)
	if !reg.MatchString(s) {
		return "", false
	}
	matchArr := reg.FindStringSubmatch(s)
	return matchArr[len(matchArr)-1], true
}
