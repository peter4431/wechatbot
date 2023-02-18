package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

func msgFilter(msg string) string {
	//replace @到下一个非空的字段 为 ''
	msg = strings.ReplaceAll(msg, " ", " ")
	log.Printf("msgFilter before %s data:%s\n", msg, base64.StdEncoding.EncodeToString([]byte(msg)))
	regex := regexp.MustCompile(`@[^\s]*`)
	ret := regex.ReplaceAllString(msg, "")
	ret = strings.TrimSpace(ret)
	log.Printf("msgFilter after %s\n", ret)
	return ret
}
func parseContent(content string) string {
	return msgFilter(content)
}
func processMessage(msg interface{}) (string, error) {
	msg = strings.TrimSpace(msg.(string))
	msgB, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	msgStr := string(msgB)

	if len(msgStr) >= 2 {
		msgStr = msgStr[1 : len(msgStr)-1]
	}

	return msgStr, nil
}
