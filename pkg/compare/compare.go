package compare

import "strings"

// CompareOutput 极其严苛又极其包容的 OJ 级字符串比对算法
func CompareOutput(userOutput string, standardAnswer string) bool {
	// strings.Fields 会极其智能地把一段长文本，按照所有的空格、换行符 \n、制表符 \t 给切碎成数组
	// 比如 "  8   \n\n  " 会被切成 ["8"]
	// 比如 "3 \n 5" 会被切成 ["3", "5"]
	userTokens := strings.Fields(userOutput)
	stdTokens := strings.Fields(standardAnswer)

	// 如果切出来的词汇数量都不一样，那绝对是错的
	if len(userTokens) != len(stdTokens) {
		return false
	}

	// 挨个比对每一个词，哪怕错了一个字符，也是 WA
	for i := range userTokens {
		if userTokens[i] != stdTokens[i] {
			return false
		}
	}

	return true
}
