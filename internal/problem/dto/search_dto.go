package dto

// SearchRequest 搜索条件 (搬到这里来)
type SearchRequest struct {
	Keyword    string   `json:"keyword"`
	Difficulty int      `json:"difficulty"`
	Tags       []string `json:"tags"`
}

// SearchResult 统一定义返回给 Controller 的结果结构
type SearchResult struct {
	Total int64                    `json:"total"`
	Data  []map[string]interface{} `json:"data"`
}
