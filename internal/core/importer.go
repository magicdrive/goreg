package core

import (
	"bytes"
	"errors"
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

	// importsMap is keyed by importKey (path and alias): the same path
	// imported under two names is two imports, and both must be kept.
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

		key := importKey(path, moduleAlias)
		importsMap[key] = model.ImportPack{
			Entity:      imp,
			Path:        path,
			LineComment: lineComment,
			Doc:         docComments,
			End:         endComment,
			Alias:       moduleAlias,
		}

		switch group {
		case model.StdLib:
			importGroupMap[model.StdLib] = append(importGroupMap[model.StdLib], key)
		case model.ThirdParty:
			importGroupMap[model.ThirdParty] = append(importGroupMap[model.ThirdParty], key)
		case model.Local:
			importGroupMap[model.Local] = append(importGroupMap[model.Local], key)
		case model.Organization:
			importGroupMap[model.Organization] = append(importGroupMap[model.Organization], key)
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
	return ReplaceImports(src, fset, node, buf.String())
}

// importKey identifies one import in FormatImports: its path, then its
// alias after a NUL byte. NUL sorts before every byte of an import path, so
// ordering keys as strings orders them by path exactly as ordering the paths
// did, and only breaks ties between imports of the same path (by alias).
func importKey(path, alias string) string {
	return path + "\x00" + alias
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
			fmt.Fprintf(buf, "\t%s \"%s\"", importPack.Alias, importPack.Path)
		} else {
			isNoneAliasImportExist = true
			fmt.Fprintf(buf, "\t\"%s\"", importPack.Path)
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

// ReplaceImports replaces the first parenthesized import declaration of
// src with newImports. file and fset are the result of parsing src: the
// declaration is located in the syntax tree, never by searching the text, so
// an "import (" or a parenthesis in a comment or a string cannot move it.
// One newline after the closing parenthesis is replaced too. Without a
// parenthesized import declaration, src is returned unchanged.
//
// The result is parsed before it is returned: when it is not valid Go, an
// error is returned instead, so that nothing broken is ever written.
func ReplaceImports(src []byte, fset *token.FileSet, file *ast.File, newImports string) ([]byte, error) {
	var decl *ast.GenDecl
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT && gd.Lparen.IsValid() {
			decl = gd
			break
		}
	}
	if decl == nil {
		return src, nil
	}

	tf := fset.File(decl.Pos())
	if tf == nil || tf.Size() != len(src) {
		return nil, errors.New("goreg: the syntax tree does not belong to the source")
	}
	start := tf.Offset(decl.Pos())
	end := tf.Offset(decl.Rparen) + 1
	if end < len(src) && src[end] == '\n' {
		end++
	}

	out := make([]byte, 0, len(src)-(end-start)+len(newImports))
	out = append(out, src[:start]...)
	out = append(out, newImports...)
	out = append(out, src[end:]...)

	if _, err := parser.ParseFile(token.NewFileSet(), "", out, parser.ParseComments); err != nil {
		return nil, fmt.Errorf("goreg: the formatted source does not parse, nothing is written: %w", err)
	}
	return out, nil
}
