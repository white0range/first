package dto

// dto 定义 (建议单独放在 models 或 dto 包，这里为了直观写在一起)
type UserAuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"` // 增加最小长度校验
}

type UserProfileResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Role        int    `json:"role"`
	SolvedCount int    `json:"solved_count"`
	SolvedList  []uint `json:"solved_list"`
}
