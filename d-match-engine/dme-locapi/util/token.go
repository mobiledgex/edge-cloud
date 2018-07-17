package util

import (
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type GDDTToken string

var (
	DefaultTokenValidSeconds = 60
)

func GenerateToken(ip string, tokenDur int) string {
	expireTime := time.Now().Local().Add(time.Second * time.Duration(tokenDur))
	tokenString := fmt.Sprintf("IP=%s Expires=%d", ip, expireTime.Unix())
	token64 := b64.StdEncoding.EncodeToString([]byte(tokenString))
	return token64
}

//Decode a token and return IP, valid, error
func DecodeToken(t GDDTToken) (string, bool, error) {

	tb, err := b64.StdEncoding.DecodeString(string(t))
	if err != nil {
		fmt.Printf("Error: %v in decoding token: %s", err, string(t))
		return "", false, err
	}
	//format is IP=ipaddr Expires=epochtime
	ts := strings.Split(string(tb), " ")
	ip := strings.Split(ts[0], "=")[1]
	expirestr := strings.Split(ts[1], "=")[1]
	expires, err := strconv.ParseInt(expirestr, 10, 64)
	if err != nil {
		fmt.Printf("Error getting expire time from token %v", err)
		return "", false, err
	}
	nowTime := time.Now().UTC().Unix()
	valid := nowTime <= expires

	fmt.Printf("decoded token ip %s expires %d now %d valid %t\n", ip, expires, nowTime, valid)
	return ip, valid, nil
}
