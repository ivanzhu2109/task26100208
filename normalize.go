package entdup

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var foldTransformer = transform.Chain(
	norm.NFKC,
	cases.Fold(),
)

// Normalize 对原始企业名称依次做 NFKC 规范化、Unicode 默认完整
// 大小写折叠，再删除所有非 Unicode 字母、非 Unicode 数字的字符。
func Normalize(raw string) string {
	folded, _, err := transform.String(foldTransformer, raw)
	if err != nil {
		folded = raw
	}
	var b strings.Builder
	b.Grow(len(folded))
	for _, r := range folded {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
