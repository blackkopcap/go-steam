package community

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

func EncryptPassword(N string, E string, password string) (string, error) {
	n, ok := new(big.Int).SetString(N, 16)
	if !ok {
		return "", errors.New("can not set string N")
	}
	e, err := strconv.ParseInt(E, 16, 32)
	if err != nil {
		return "", err
	}
	rsaPubKey := rsa.PublicKey{
		N: n,
		E: int(e),
	}
	encryptedPassword, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPubKey, []byte(password))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedPassword), nil
}

func GenerateSessionID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateDeviceID(accountName string, password string) string {
	sum := md5.Sum([]byte(accountName + password))
	deviceID := fmt.Sprintf(
		"android:%x-%x-%x-%x-%x",
		sum[:2], sum[2:4], sum[4:6], sum[6:8], sum[8:10],
	)
	return deviceID
}

func SetCookies(client *http.Client, cookies []*http.Cookie) error {
	if client.Jar == nil {
		jar, err := cookiejar.New(new(cookiejar.Options))
		if err != nil {
			return err
		}
		client.Jar = jar
	}
	communityUrl, err := url.Parse(baseUrl)
	if err != nil {
		return err
	}
	client.Jar.SetCookies(communityUrl, cookies)
	return nil
}
