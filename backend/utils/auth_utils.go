package utils

import (
	"backend/global"
	"backend/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//哈希辅助函数
func Hpwd(pwd string) (string, error) {
	hpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hpwd), err
}
//JWT生成函数
func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	signedtoken, err := token.SignedString([]byte("secret"))
	return "Bearer " + signedtoken, err
}

// GenerateJWTWithRole 生成包含 role 的 JWT
// GenerateJWTWithRole 生成包含 role 的 JWT，若 merchantId 非空则把 merchantId 一并放入 claims
func GenerateJWTWithRole(username, role string, merchantId string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	if merchantId != "" {
		claims["merchantId"] = merchantId
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedtoken, err := token.SignedString([]byte("secret"))
	return "Bearer " + signedtoken, err
}

func CheckPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ParseJWT(tokenstring string) (string, error) {
	if len(tokenstring) > 7 && tokenstring[:7] == "Bearer " {
		tokenstring = tokenstring[7:]
	}
	token, err := jwt.Parse(tokenstring, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("invalid username")
		}
		return username, nil
	}
	return "", errors.New("invalid token")
}

// get hashpassword
func GetUserHashByUsernameuser(username string) (string, error) {
	var user models.BaseUser
	result := global.Db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", result.Error
	}
	return user.Password, nil
}

// updateUserPasswordHash 根据用户名更新密码哈希
func UpdateUserPasswordHash(username, newHash string) error {
	return global.Db.Model(&models.BaseUser{}).Where("username = ?", username).Update("password", newHash).Error
}
