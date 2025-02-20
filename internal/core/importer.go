package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

func FormatImports(src []byte, modulePath string) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	importsMap := make(map[string]ImportPack)
	var stdLibImports, thirdPartyImports, localImports []string

	LineComments := ExtractLineComments(node, fset)

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		group := GetImportGroup(path, modulePath)

		docComments, endComment, moduleAlias := ExtractComments(imp)

		line := fset.Position(imp.Pos()).Line
		preComment := LineComments[line]

		importsMap[path] = ImportPack{
			Entity:      imp,
			LineComment: preComment,
			Doc:         docComments,
			End:         endComment,
			Alias:       moduleAlias,
		}

		switch group {
		case stdLib:
			stdLibImports = append(stdLibImports, path)
		case thirdParty:
			thirdPartyImports = append(thirdPartyImports, path)
		case local:
			localImports = append(localImports, path)
		}
	}

	sortImports(stdLibImports, importsMap)
	sortImports(thirdPartyImports, importsMap)
	sortImports(localImports, importsMap)

	var buf bytes.Buffer
	buf.WriteString("import (\n")

	groups := [][]string{}
	if len(stdLibImports) > 0 {
		groups = append(groups, stdLibImports)
	}
	if len(thirdPartyImports) > 0 {
		groups = append(groups, thirdPartyImports)
	}
	if len(localImports) > 0 {
		groups = append(groups, localImports)
	}

	for i, group := range groups {
		isLastGroup := (i == len(groups)-1)
		WriteImports(fset, &buf, group, importsMap, isLastGroup)
	}

	buf.WriteString(")\n")

	return ReplaceImports(src, buf.String()), nil
}

func GetImportGroup(pkg string, modulePath string) ImportGroup {
	if strings.HasPrefix(pkg, modulePath) {
		return local
	}
	if !strings.Contains(pkg, ".") {
		return stdLib
	}
	return thirdParty
}

func sortImports(imports []string, importsMap map[string]ImportPack) {
	sort.SliceStable(imports, func(i, j int) bool {
		aliasI := importsMap[imports[i]].Alias
		aliasJ := importsMap[imports[j]].Alias

		if aliasI == "" && aliasJ != "" {
			return true
		}
		if aliasI != "" && aliasJ == "" {
			return false
		}
		return imports[i] < imports[j]
	})
}

func WriteImports(fset *token.FileSet, buf *bytes.Buffer, pkgs []string, importsMap map[string]ImportPack, isLastGroup bool) {
	isFirstImport := true
	isNoneAliasImport := true
	isNoneAliasImportExist := false

	for _, imp := range pkgs {
		importPack, exists := importsMap[imp]
		if !exists {
			fmt.Fprintf(buf, "\t\"%s\"\n", imp)
			continue
		}

		if !isFirstImport && len(importPack.Doc) > 0 {
			buf.WriteString("\n")
		}
		isFirstImport = false

		for _, c := range importPack.Doc {
			buf.WriteString(fmt.Sprintf("\t%s\n", c))
		}

		if importPack.LineComment != nil {
			if IsCommentBeforeImport(fset, importPack.Entity, importPack.LineComment) {
				buf.WriteString(fmt.Sprintf("\t%s\n", importPack.LineComment.Text))
			}
		}

		if importPack.Alias != "" {
			if isNoneAliasImport && isNoneAliasImportExist {
				buf.WriteString("\n")
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
	}

	if !isLastGroup {
		buf.WriteString("\n")
	}
}

func ExtractComments(imp *ast.ImportSpec) ([]string, string, string) {
	var docComments []string
	var endComment string
	var alias string = ""

	if imp.Doc != nil {
		for _, c := range imp.Doc.List {
			docComments = append(docComments, c.Text)
		}
	}
	if imp.Comment != nil && len(imp.Comment.List) > 0 {
		endComment = strings.TrimSpace(imp.Comment.List[0].Text)
	}

	if imp.Name != nil {
		alias = imp.Name.Name
	}

	return docComments, endComment, alias
}

func ExtractLineComments(node *ast.File, fset *token.FileSet) map[int]*ast.Comment {
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

func IsCommentBeforeImport(fset *token.FileSet, imp *ast.ImportSpec, comment *ast.Comment) bool {
	commentPos := fset.Position(comment.Pos()).Offset

	importPos := fset.Position(imp.Pos()).Offset

	return commentPos < importPos
}

func ReplaceImports(src []byte, newImports string) []byte {
	srcStr := string(src)
	start := strings.Index(srcStr, "import (")
	if start == -1 {
		return src
	}

	end := start
	bracketCount := 1
	for i := start + len("import ("); i < len(srcStr); i++ {
		if srcStr[i] == '(' {
			bracketCount++
		} else if srcStr[i] == ')' {
			bracketCount--
			if bracketCount == 0 {
				end = i + 1
				break
			}
		}
	}

	if end < len(srcStr) && srcStr[end] == '\n' {
		end++
	}

	return []byte(srcStr[:start] + newImports + srcStr[end:])
}
