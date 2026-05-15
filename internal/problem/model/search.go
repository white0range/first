package model

// EsProblem 定义存入 ES 的数据结构
// 💡 架构师细节：存入 ES 的数据不需要像 MySQL 那么全！
// 我们只存“需要被搜索”的字段（比如标题、描述、标签），以及能定位回 MySQL 的主键 ID。
type EsProblem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	//Difficulty  int    `json:"difficulty"` // 假设你的难度是 1,2,3 这种数字
	Tags []string `json:"tags"` // 如果你有标签，可以拼成字符串丢进来
}
