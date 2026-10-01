package companydedup

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var foldCase = cases.Fold()

// normalize 对名称依次做 NFKC 规范化、Unicode 默认完整大小写折叠，
// 并删除所有非 Unicode 字母、非 Unicode 数字的字符。
func normalize(name string) string {
	nfkc := norm.NFKC.String(name)
	folded := foldCase.String(nfkc)
	var b strings.Builder
	b.Grow(len(folded))
	for _, r := range folded {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
