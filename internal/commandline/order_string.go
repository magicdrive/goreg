package commandline

import (
	"fmt"
	"strings"

	"github.com/magicdrive/goreg/internal/model"
)

var wordMap = map[string]model.ImportGroup{
	"std":          model.StdLib,
	"stdlib":       model.StdLib,
	"s":            model.StdLib,
	"thirdparty":   model.ThirdParty,
	"third_party":  model.ThirdParty,
	"3rdparty":     model.ThirdParty,
	"3rd_party":    model.ThirdParty,
	"3rd":          model.ThirdParty,
	"3":            model.ThirdParty,
	"t":            model.ThirdParty,
	"local":        model.Local,
	"l":            model.Local,
	"organization": model.Organization,
	"org":          model.Organization,
	"o":            model.Organization,
}

func FilterValidWords(input string) ([]model.ImportGroup, error) {
	result := make([]model.ImportGroup, 0, 16)
	var sb strings.Builder

	for i := 0; i < len(input); i++ {
		c := input[i]
		if c == ',' {
			word := sb.String()
			if id, exists := wordMap[word]; !exists {
				return nil, fmt.Errorf("specified for --order is invalid.: %s", word)
			} else {
				result = append(result, id)
			}
			sb.Reset()
		} else if c != ' ' {
			sb.WriteByte(c)
		}
	}

	if sb.Len() > 0 {
		word := sb.String()
		if id, exists := wordMap[word]; !exists {
			return nil, fmt.Errorf("specified for --order is invalid.: %s", word)
		} else {
			result = append(result, id)
		}
	}

	return result, nil
}

func GenerateOrderStrings(input string) ([]model.ImportGroup, error) {
	validOrder, err := FilterValidWords(input)
	if err != nil {
		return nil, err
	}
	result := Unique(validOrder)

	if len(result) != 4 {
		return nil, fmt.Errorf("--order must include all of std, thirdparty, organization, and local.")
	} else {
		return result, nil
	}
}
