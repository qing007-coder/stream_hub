package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 对用户密码进行哈希（存数据库用）
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost, // 工业级默认
	)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword 校验用户输入密码是否正确
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
	return err == nil
}
