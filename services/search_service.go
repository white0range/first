package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	// 记得替换成你实际的 module 名字
	"gojo/global"
	"gojo/models"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

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

// ExecuteSearch 执行核心搜索逻辑 (只认参数，不认 HTTP)
func ExecuteSearch(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	if req.Keyword != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Keyword,
				"fields": []string{"title^3", "description"},
			},
		})
	}
	//if req.Difficulty != 0 {
	//	filter = append(filter, map[string]interface{}{
	//		"term": map[string]interface{}{"difficulty": req.Difficulty},
	//	})
	//}
	if len(req.Tags) > 0 {
		filter = append(filter, map[string]interface{}{
			"terms": map[string]interface{}{"tags.keyword": req.Tags}, //加上 .keyword！告诉 ES 用那个没被切碎的备份账本！
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{"must": must, "filter": filter},
		},
		"size": 50,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := global.EsClient.Search(
		global.EsClient.Search.WithContext(ctx),
		global.EsClient.Search.WithIndex("problems"),
		global.EsClient.Search.WithBody(&buf),
		global.EsClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES 返回错误: %s", res.String())
	}

	var esResult map[string]interface{}
	json.NewDecoder(res.Body).Decode(&esResult)

	hits := esResult["hits"].(map[string]interface{})["hits"].([]interface{})
	var resultData []map[string]interface{}

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		source["_score"] = hit.(map[string]interface{})["_score"]
		resultData = append(resultData, source)
	}

	total := int64(esResult["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	return &SearchResult{
		Total: total,
		Data:  resultData,
	}, nil
}

// SyncAllProblemsToES 全量同步 MySQL 数据到 ES (老板发话：把仓库搬空！)
func SyncAllProblemsToES() {
	ctx := context.Background()
	var problems []models.Problem

	// 2. 加上 Preload 魔法，连同标签一起查！
	if err := models.DB.Preload("Tags").Find(&problems).Error; err != nil {
		log.Printf("❌ 数据库查询失败: %v\n", err)
		return
	}

	fmt.Printf("📦 准备将 %d 道题目同步至 Elasticsearch...\n", len(problems))
	successCount := 0

	// 2. 遍历每一道题，打包发给 ES
	for _, p := range problems {
		// 3. 把 GORM 查出来的 []Tag 结构体，转换成纯字符串数组 []string
		var tagNames []string
		for _, t := range p.Tags {
			tagNames = append(tagNames, t.Name)
		}
		// 组装要在 ES 里建立索引的数据
		doc := EsProblem{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			Tags:        tagNames,
			// Difficulty: p.Difficulty, // 按你实际的 Problem 结构体字段来赋值
		}

		// 把 Go 结构体序列化成 JSON 字节流
		body, _ := json.Marshal(doc)

		// 3. 构建发送给 ES 的请求 (极其核心的一步)
		req := esapi.IndexRequest{
			Index:      "problems",              // 告诉 ES，存在哪个账本（表）里
			DocumentID: strconv.Itoa(int(p.ID)), // 🚨 极其关键：让 ES 的 ID 和 MySQL 的 ID 绝对绑定！
			Body:       bytes.NewReader(body),
			Refresh:    "true", // 存完立刻刷新，保证下一秒就能搜到
		}

		// 执行发送
		res, err := req.Do(ctx, global.EsClient)
		if err != nil {
			log.Printf("⚠️ 题目 ID %d 同步因网络异常失败: %v\n", p.ID, err)
			continue
		}

		// 检查 ES 返回的状态码是不是报错了
		if res.IsError() {
			log.Printf("⚠️ 题目 ID %d 被 ES 拒绝，返回详情: %s\n", p.ID, res.String())
		} else {
			successCount++
		}

		// 用完必须关闭连接池，不然内存泄漏！
		if res != nil {
			res.Body.Close()
		}
	}

	fmt.Printf("✅ ES 数据初始化完毕！共成功注入 %d 道题目数据！\n", successCount)
}
