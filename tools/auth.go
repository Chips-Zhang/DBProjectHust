package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/Chips-zhang/DBProjectHust/service"
)

func PasswordSaltedHash(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + password + salt))

	return hex.EncodeToString(h.Sum(nil))
}

func CheckPassword(id UidT, password string) error {
	u := UserInfo{
		Id: id,
	}
	err := DB_.Select(&u)
	if err != nil {
		return err
	}

	if u.Password != password {
		return errors.New("Invalid password.")
	}
	return nil
}

var tokenUidMap map[string]UidT

func InitAuthModule() {
	tokenUidMap = make(map[string]UidT)
	rand.Seed(time.Now().UnixNano())
}

func GenerateToken() string {
	timestamp := time.Now().Unix() + rand.Int63()
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(strconv.FormatInt(timestamp, 10))))[:45]
	if _, ok := tokenUidMap[token]; ok {
		return GenerateToken()
	} else {
		return token
	}
}

func DoLogin(username, password string) (string, error) {
	u, err2 := UsernameToInfo(username)
	if err2 != nil {
		return "", err2
	}

	if u.Password != password {
		return "", errors.New("Password denied.")
	}

	token := GenerateToken()
	tokenUidMap[token] = u.Id
	return token, nil
}

func DoLogout(token string) error {
	if _, ok := tokenUidMap[token]; ok {
		delete(tokenUidMap, token)
		return nil
	} else {
		return errors.New("Invalid token")
	}
}

func VerifyToken(token string) (UidT, error) {
	if val, ok := tokenUidMap[token]; ok {
		return val, nil
	} else {
		return -1, errors.New("Invalid token")
	}
}

var domainRegex = regexp.MustCompile("^[a-zA-Z0-9:./_-]*$")

func ForgetPassword(email, domain, proto string) error {
	var u UserInfo
	if !UsernameRegex.MatchString(email) {
		return errors.New("Invalid email format.")
	}
	err := DB_.Model(&u).Where("email = '" + email + "'").Select()
	if err != nil {
		if err.Error() == PgNotFoundErr {
			return errors.New("Email not found.")
		}
		return err
	}

	if !domainRegex.MatchString(domain) {
		return errors.New("Invalid domain format.")
	}

	emailTxt := fmt.Sprintf("Your username is %s. Use the following link to reset your password:\n %s//%s/changePassword.html?old=%s&name=%s\n\nTMobile",
		u.Name, proto, domain, u.Password, u.Name)
	go func() {
		err := service.SendEmail(email, "TMobile password reset", emailTxt)
		if err == nil {
			log.Print("Sent email to " + email)
		} else {
			log.Print("Unable to send email to " + email + ", " + err.Error())
		}
	}()

	return nil
}

func ChangePassword(name, old, new string) error {
	u, err := UsernameToInfo(name)
	if err != nil {
		return err
	}

	if u.Password != old {
		return errors.New("Invalid old password.")
	}
	u.Password = new

	err2 := DB_.Update(&u)
	return err2
}
