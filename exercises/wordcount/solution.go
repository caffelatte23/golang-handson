//go:build solution

package wordcount

import "strings"

// WordCount は単語ごとの出現回数を返します。
func WordCount(s string) map[string]int {
	counts := make(map[string]int)
	for _, w := range strings.Fields(s) {
		counts[w]++ // 存在しないキーはゼロ値 0 から始まる
	}
	return counts
}
