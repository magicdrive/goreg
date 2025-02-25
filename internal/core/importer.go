package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"

	"github.com/magicdrive/goreg/internal/commandline"
	"github.com/magicdrive/goreg/internal/model"
)

func FormatImports(src []byte, opt *commandline.Option) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	importsMap := make(map[string]model.ImportPack)
	importGroupMap := map[model.ImportGroup][]string{
		model.StdLib:       {},
		model.ThirdParty:   {},
		model.Local:        {},
		model.Organization: {},
	}

	lineComments := ExtractLineComments(node, fset, opt)

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		group := GetImportGroup(path, opt)

		docComments, endComment, moduleAlias := ExtractComments(imp, opt)
		line := fset.Position(imp.Pos()).Line
		lineComment := lineComments[line]

		importsMap[path] = model.ImportPack{
			Entity:      imp,
			LineComment: lineComment,
			Doc:         docComments,
			End:         endComment,
			Alias:       moduleAlias,
		}

		switch group {
		case model.StdLib:
			importGroupMap[model.StdLib] = append(importGroupMap[model.StdLib], path)
		case model.ThirdParty:
			importGroupMap[model.ThirdParty] = append(importGroupMap[model.ThirdParty], path)
		case model.Local:
			importGroupMap[model.Local] = append(importGroupMap[model.Local], path)
		case model.Organization:
			importGroupMap[model.Organization] = append(importGroupMap[model.Organization], path)
		}
	}

	sortImports(importGroupMap[model.StdLib], importsMap, opt)
	sortImports(importGroupMap[model.ThirdParty], importsMap, opt)
	sortImports(importGroupMap[model.Local], importsMap, opt)
	sortImports(importGroupMap[model.Organization], importsMap, opt)

	var buf bytes.Buffer
	buf.WriteString("import (\n")

	groups := [][]string{}

	for _, elem := range opt.ImportOrder {
		if len(importGroupMap[elem]) > 0 {
			groups = append(groups, importGroupMap[elem])
		}
	}

	for i, group := range groups {
		isLastGroup := (i == len(groups)-1)
		WriteImports(fset, &buf, group, importsMap, opt, isLastGroup)
	}

	buf.WriteString(")\n")
	return ReplaceImports(src, buf.String()), nil
}

func GetImportGroup(pkg string, opt *commandline.Option) model.ImportGroup {
	if strings.HasPrefix(pkg, opt.ModulePath) {
		return model.Local
	}
	if opt.OrganizationName != "" && strings.HasPrefix(pkg, opt.OrganizationName) {
		return model.Organization
	}
	if !strings.Contains(pkg, ".") {
		return model.StdLib
	}
	return model.ThirdParty
}

func sortImports(imports []string, importsMap map[string]model.ImportPack, opt *commandline.Option) {
	var sortArgo func(i, j int) bool
	if opt.SortIncludeAliasFlag {
		sortArgo = func(i, j int) bool {
			return imports[i] < imports[j]
		}
	} else {
		sortArgo = func(i, j int) bool {
			iData, jData := importsMap[imports[i]], importsMap[imports[j]]
			iAlias, jAlias := iData.Alias != "", jData.Alias != ""

			if iAlias && !jAlias {
				return false
			}
			if !iAlias && jAlias {
				return true
			}
			return imports[i] < imports[j]
		}

	}
	sort.SliceStable(imports, sortArgo)
}

func WriteImports(fset *token.FileSet, buf *bytes.Buffer, pkgs []string,
	importsMap map[string]model.ImportPack, opt *commandline.Option, isLastGroup bool) {
	isFirstImport := true
	isNoneAliasImport := true
	isNoneAliasImportExist := false

	for _, imp := range pkgs {
		importPack := importsMap[imp]
		lineBreaked := false

		if !isFirstImport && len(importPack.Doc) > 0 {
			buf.WriteString("\n")
			lineBreaked = true
		}

		for _, c := range importPack.Doc {
			buf.WriteString(fmt.Sprintf("\t%s\n", c))
		}

		if importPack.LineComment != nil && IsCommentBeforeImport(importPack.Entity, importPack.LineComment) {
			if !lineBreaked && !isFirstImport {
				buf.WriteString("\n")
				lineBreaked = true
			}
			buf.WriteString(fmt.Sprintf("\t%s\n", importPack.LineComment.Text))
		}

		if importPack.Alias != "" {
			if isNoneAliasImport && isNoneAliasImportExist && !opt.MinimizeGroupFlag {
				if !lineBreaked {
					buf.WriteString("\n")
					lineBreaked = true
				}
				isNoneAliasImport = false
			}
			fmt.Fprintf(buf, "\t%s \"%s\"", importPack.Alias, imp)
		} else {
			isNoneAliasImportExist = true
			fmt.Fprintf(buf, "\t\"%s\"", imp)
		}

		if importPack.End != "" {
			buf.WriteString(" " + importPack.End)
		}
		buf.WriteString("\n")
		isFirstImport = false
	}

	if !isLastGroup {
		buf.WriteString("\n")
	}
}

func ExtractComments(imp *ast.ImportSpec, opt *commandline.Option) ([]string, string, string) {
	var docComments []string
	var endComment, alias string

	if imp.Doc != nil && len(imp.Doc.List) > 0 && !opt.RemoveImportCommentFlag {
		for _, c := range imp.Doc.List {
			docComments = append(docComments, c.Text)
		}
	}
	if imp.Comment != nil && len(imp.Comment.List) > 0 && !opt.RemoveImportCommentFlag {
		endComment = strings.TrimSpace(imp.Comment.List[0].Text)
	}
	if imp.Name != nil {
		alias = imp.Name.Name
	}

	return docComments, endComment, alias
}

func ExtractLineComments(node *ast.File, fset *token.FileSet, opt *commandline.Option) map[int]*ast.Comment {

	if opt.RemoveImportCommentFlag {
		return map[int]*ast.Comment{}
	}

	precedingComments := make(map[int]*ast.Comment)

	for _, commentGroup := range node.Comments {
		for _, comment := range commentGroup.List {
			nextPos := comment.Pos() + 1
			line := fset.Position(nextPos).Line
			precedingComments[line] = comment
		}
	}
	return precedingComments
}

func IsCommentBeforeImport(imp *ast.ImportSpec, comment *ast.Comment) bool {
	return comment.Pos() < imp.Pos()
}

func ReplaceImports(src []byte, newImports string) []byte {
	srcStr := string(src)
	start := bytes.Index(src, []byte("import ("))
	if start == -1 {
		return src
	}

	end := start
	bracketCount := 1

	// len("import (") is 8
	for i := start + 8; i < len(srcStr); i++ {
		switch srcStr[i] {
		case '(':
			bracketCount++
		case ')':
			bracketCount--
			if bracketCount == 0 {
				end = i + 1
				goto replace
			}
		}
	}

replace:
	if end < len(srcStr) && srcStr[end] == '\n' {
		end++
	}

	var builder strings.Builder
	builder.Grow(len(srcStr) + len(newImports))
	builder.WriteString(srcStr[:start])
	builder.WriteString(newImports)
	builder.WriteString(srcStr[end:])

	return []byte(builder.String())
}
