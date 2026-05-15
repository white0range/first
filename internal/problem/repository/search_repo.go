package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gojo/infrastructure/search"  // 你的 ES 客户端初始化包
	"gojo/internal/problem/model" // 假设 EsProblem 在这里
	"strconv"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// 1. 定义 ES 专属仓管接口
type ProblemSearchRepository interface {
	SearchProblems(ctx context.Context, keyword string, tags []string) (int64, []map[string]interface{}, error)
	UpsertProblemToES(ctx context.Context, doc model.EsProblem) error
}

type problemSearchRepoES struct{}

func NewProblemSearchRepository() ProblemSearchRepository {
	return &problemSearchRepoES{}
}

// 2. 搜索落地实现 (把拼接 JSON 和解析 Hits 全部封死在这里)
func (r *problemSearchRepoES) SearchProblems(ctx context.Context, keyword string, tags []string) (int64, []map[string]interface{}, error) {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	if keyword != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  keyword,
				"fields": []string{"title^3", "description"},
			},
		})
	}
	if len(tags) > 0 {
		filter = append(filter, map[string]interface{}{
			"terms": map[string]interface{}{"tags.keyword": tags},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{"bool": map[string]interface{}{"must": must, "filter": filter}},
		"size":  50,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := search.EsClient.Search(
		search.EsClient.Search.WithContext(ctx),
		search.EsClient.Search.WithIndex("problems"),
		search.EsClient.Search.WithBody(&buf),
		search.EsClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, nil, fmt.Errorf("ES 返回错误: %s", res.String())
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

	return total, resultData, nil
}

// 3. 写入 ES 落地实现
func (r *problemSearchRepoES) UpsertProblemToES(ctx context.Context, doc model.EsProblem) error {
	body, _ := json.Marshal(doc)
	req := esapi.IndexRequest{
		Index:      "problems",
		DocumentID: strconv.Itoa(int(doc.ID)),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, search.EsClient)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES 拒绝写入: %s", res.String())
	}
	return nil
}
