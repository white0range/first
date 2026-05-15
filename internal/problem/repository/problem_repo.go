package repository

import (
	"context"
	"gojo/infrastructure/mysql"
	"gojo/internal/problem/model"
	model2 "gojo/internal/submission/model"
	"gojo/pkg/pagination"

	"gorm.io/gorm"
)

// 1. 定义仓管的接口 (大厂铁律：面向接口编程)
type ProblemRepository interface {
	CreateProblem(ctx context.Context, problem *model.Problem) error
	GetProblemByID(ctx context.Context, id string) (*model.Problem, error)
	// 获取分页题目数据
	GetProblemList(ctx context.Context, offset, limit int, tagID string) ([]model.Problem, int64, error)
	// 获取用户对指定题目组的 AC 状态
	GetUserACProblemIDs(ctx context.Context, uid uint, problemIDs []uint) ([]uint, error)
	UpdateProblem(ctx context.Context, problemID string, data map[string]interface{}) error
	DeleteProblem(ctx context.Context, id string) error
	UpdateProblemTags(ctx context.Context, id string, tagIDs []uint) error
	// 在 ProblemRepository 接口里增加：
	GetAllProblemsWithTags(ctx context.Context) ([]model.Problem, error)
	GetTagsByIDs(ctx context.Context, ids []uint) ([]model.Tag, error)
}

// 2. 具体的实现类 (MySQL 仓管)
type problemRepoMysql struct{}

func NewProblemRepository() ProblemRepository {
	return &problemRepoMysql{}
}

// CreateProblem 仓管只负责把拼装好的对象塞进数据库
func (r *problemRepoMysql) CreateProblem(ctx context.Context, problem *model.Problem) error {
	// GORM 的 Create 自带事务，一行搞定级联插入
	return mysql.DB.WithContext(ctx).Create(problem).Error
}
func (r *problemRepoMysql) GetTagsByIDs(ctx context.Context, ids []uint) ([]model.Tag, error) {
	var tags []model.Tag
	err := mysql.DB.WithContext(ctx).Find(&tags, ids).Error
	return tags, err
}

// GetProblemByID 仓管只负责去 MySQL 里捞数据
func (r *problemRepoMysql) GetProblemByID(ctx context.Context, id string) (*model.Problem, error) {
	var problem model.Problem
	err := mysql.DB.WithContext(ctx).Preload("Tags").First(&problem, id).Error
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepoMysql) GetProblemList(ctx context.Context, page, limit int, tagID string) ([]model.Problem, int64, error) {
	var items []model.Problem
	var total int64
	// 👈 修改开始：构建动态的 GORM 查询器
	query := mysql.DB.WithContext(ctx).Model(&model.Problem{})
	// 如果前端传了 tag_id，立刻启动 JOIN 联表过滤！
	if tagID != "" {
		query = query.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
			Where("problem_tags.tag_id = ?", tagID)
	}

	// 3. 先查总数 (注意：Count的时候千万不能带 limit 和 offset)
	//models.DB.Model(&models.Problem{}).Count(&total)
	query.Count(&total)
	// 💥 弃用 utils.Paginate(c)，在 Service 层手动计算 Offset（极其简单）
	//offset := (res.Page - 1) * res.Limit
	// 【大厂性能优化点】：Select
	// 列表页通常只需要展示 ID、标题、通过率。
	// 题目的 Description（描述）通常是一大段长文本，如果列表页也全查出来，会极大地浪费服务器带宽！
	// 所以我们用 Select 指定只捞取需要的列，这就是 GORM 的优雅之处。
	//models.DB.Select("id", "title", "submit_count", "accepted_count").
	//Limit(limit).
	//Offset(offset).
	//Find(&problems)

	// 💥 见证奇迹的时刻：干掉繁琐的 Offset 计算，直接挂载分页插件！
	// ⚠️ 极其关键：因为用了 JOIN，Select 里的字段必须加上 "problems." 前缀，否则 MySQL 会报 Ambiguous column name (字段歧义)
	err := query.Select("problems.id", "problems.title", "problems.submit_count", "problems.accepted_count").
		Preload("Tags"). // 👈 新增：把题目关联的标签小尾巴也带上发给前端！
		Scopes(pagination.Paginate(page, limit)).
		Find(&items).Error
	// 👈 修改结束
	return items, total, err
}

func (r *problemRepoMysql) GetUserACProblemIDs(ctx context.Context, uid uint, problemIDs []uint) ([]uint, error) {
	var userACList []uint
	err := mysql.DB.WithContext(ctx).Model(&model2.Submission{}).
		Where("user_id = ? AND status = 'AC' AND problem_id IN ?", uid, problemIDs).
		Distinct("problem_id").
		Pluck("problem_id", &userACList).Error
	return userACList, err
}

func (r *problemRepoMysql) UpdateProblem(ctx context.Context, problemID string, data map[string]interface{}) error {
	// GORM 的 Create 自带事务，一行搞定级联插入
	return mysql.DB.WithContext(ctx).Model(&model.Problem{}).Where("id = ?", problemID).Updates(data).Error
}

// 2. 在实现类中落地（🚨 注意这里的骚操作：使用 tx 事务代替普通的 db 调用）
func (r *problemRepoMysql) DeleteProblem(ctx context.Context, id string) error {
	// 使用 GORM 的 Transaction 方法，只要返回 err，里面的所有 SQL 自动撤销！
	return mysql.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查验是否存在
		var problem model.Problem
		if err := tx.First(&problem, id).Error; err != nil {
			return err // 找不到，直接打断事务，抛出 ErrRecordNotFound
		}

		// 2. 斩草除根：干掉关联的测试用例
		if err := tx.Where("problem_id = ?", id).Delete(&model.TestCase{}).Error; err != nil {
			return err
		}

		// 3. 🚨 满足你的需求：清理题目标签关联表！
		// 极其优雅的 GORM 魔法：Clear() 会自动去 problem_tags 中间表里，删掉所有 problem_id 等于当前 id 的记录，但绝对不会误删标签库里的真实标签！
		if err := tx.Model(&problem).Association("Tags").Clear(); err != nil {
			return err
		}

		// 4. 最后：把题目本身干掉
		if err := tx.Delete(&problem).Error; err != nil {
			return err
		}

		return nil // 完美执行，提交事务！
	})
}

func (r *problemRepoMysql) UpdateProblemTags(ctx context.Context, id string, tagIDs []uint) error {
	// 使用 GORM 的 Transaction 方法，只要返回 err，里面的所有 SQL 自动撤销！
	return mysql.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var problem model.Problem
		// 1. 查验是否存在
		if err := tx.First(&problem, id).Error; err != nil {
			return err // 找不到，直接打断事务，抛出 ErrRecordNotFound
		}
		// 2. 把前端传来的新标签 ID，去数据库换成真实的 Tag 实体
		var tags []model.Tag
		if len(tagIDs) > 0 {
			if err := tx.Find(&tags, tagIDs).Error; err != nil {
				return err
			}
		}

		// 3. 💥 GORM 魔法：关联替换 (Replace)
		if err := tx.Model(&problem).Association("Tags").Replace(tags); err != nil {
			return err
		}

		return nil // 完美执行，提交事务！
	})
}

// 实现：
func (r *problemRepoMysql) GetAllProblemsWithTags(ctx context.Context) ([]model.Problem, error) {
	var problems []model.Problem
	// 没有任何 limit 的全表扫描（由于是后台管理员偶尔执行，这里直接查可以接受）
	err := mysql.DB.WithContext(ctx).Preload("Tags").Find(&problems).Error
	return problems, err
}
