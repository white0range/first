package password

import "golang.org/x/crypto/bcrypt"

// HashPassword 负责把明文密码变成加密后的乱码
func HashPassword(password string) (string, error) {
	// GenerateFromPassword 会帮我们自动加盐。第二个参数是加密强度，通常填 12 左右即可
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPasswordHash 负责对比：用户输入的明文 vs 数据库里的加密乱码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // 如果 err 为空，说明对上了
}
