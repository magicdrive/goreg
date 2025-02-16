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

	importsMap := make(map[string]ImportBlock)
	var stdLibImports, thirdPartyImports, localImports []string

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		group := GetImportGroup(path, modulePath)

		docComments, endComment, moduleAlias := ExtractComments(imp)
		importsMap[path] = ImportBlock{
			Doc:   docComments,
			End:   endComment,
			Alias: moduleAlias,
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
		WriteImports(&buf, group, importsMap, isLastGroup)
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

func sortImports(imports []string, importsMap map[string]ImportBlock) {
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

func WriteImports(buf *bytes.Buffer, pkgs []string, importsMap map[string]ImportBlock, isLastGroup bool) {
	isFirstImport := true
	isNoneAliasImport := true
	for _, imp := range pkgs {
		comments, exists := importsMap[imp]
		if !exists {
			fmt.Fprintf(buf, "\t\"%s\"\n", imp)
			continue
		}

		if !isFirstImport && len(comments.Doc) > 0 {
			buf.WriteString("\n")
		}
		isFirstImport = false

		for _, c := range comments.Doc {
			buf.WriteString(fmt.Sprintf("\t%s\n", c))
		}

		if comments.Alias != "" {
			if isNoneAliasImport {
				buf.WriteString("\n")
				isNoneAliasImport = false
			}
			fmt.Fprintf(buf, "\t%s \"%s\"", comments.Alias, imp)
		} else {
			fmt.Fprintf(buf, "\t\"%s\"", imp)
		}

		if comments.End != "" {
			buf.WriteString(" " + comments.End)
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
