package util

import (
	"encoding/json"
	"errors"
	"github.com/bincooo/requests"
	"github.com/bincooo/requests/models"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type smailEndpoint struct {
	Key       string  `json:"rapidapi_key"`
	Endpoint  string  `json:"rapidapi_endpoint"`
	Email     string  `json:"-"`
	Timestamp float64 `json:"-"`
}

// 获取smailpro的JSID
func smailAdvancedAppJSID(proxy string) (string, error) {
	// https://smailpro.com/advanced
	response, _, err := newRequest(5*time.Second, proxy, http.MethodGet, string(ED)+"advanced", nil, nil)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", errors.New("advanced.html::405 Not Allowed")
	}
	compileRegex := regexp.MustCompile(`app.js\?v=\S+"`)
	matchStr := compileRegex.FindString(response.Text)
	if len(matchStr) == 0 {
		return "", errors.New("create_email error::not match advanced.html")
	}

	id := strings.TrimSuffix(strings.TrimPrefix(matchStr, "app.js?v="), "\"")
	return id, nil
}

func smailAppV2Id(jsId, proxy string) (string, error) {
	params := map[string]any{
		"v": jsId,
	}
	// https://smailpro.com/js/app.js
	response, _, err := newRequest(5*time.Second, proxy, http.MethodGet, string(ED)+"js/app.js", params, nil)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", errors.New("smailpro_v2_email.js::405 Not Allowed")
	}
	compileRegex := regexp.MustCompile(`\S:"smailpro_v2_email"`)
	matchStr := compileRegex.FindString(response.Text)
	if len(matchStr) == 0 {
		return "", errors.New("create_email error::not match smailpro_v2_email")
	}
	split := strings.Split(matchStr, ":")
	v2Regex := regexp.MustCompile(split[0] + `:"\w+"`)
	v2Slice := v2Regex.FindAllString(response.Text, -1)
	if len(v2Slice) < 2 {
		return "", errors.New("create_email error::not match smailpro_v2_email")
	}

	id := strings.Replace(strings.Split(v2Slice[1], ":\"")[1], "\"", "", -1)
	return id, nil
}

// 获取smailpro的端点api
func smailEndpointKey(v2Id, proxy string) (*smailEndpoint, *requests.Session, error) {
	params := map[string]any{
		"id": v2Id,
	}
	response, session, err := newRequest(5*time.Second, proxy, http.MethodGet, string(ED)+"js/chunks/smailpro_v2_email.js", params, nil)
	if err != nil {
		return nil, nil, err
	}
	if response.StatusCode != 200 {
		return nil, nil, errors.New("smailpro_v2_email.js::405 Not Allowed")
	}
	compileRegex := regexp.MustCompile(`{\s*rapidapi_endpoint:[^}]+}[\s\n]*}`)
	matchStr := compileRegex.FindString(response.Text)
	if len(matchStr) == 0 {
		return nil, nil, errors.New("create_email error::not match rapidapi_endpoint")
	}

	compileRegex = regexp.MustCompile(`([^",{][^(http)][a-z_A-Z]\w*[^"]):`)
	regStr := compileRegex.ReplaceAllString(matchStr, `"$1":`)
	var endpoint smailEndpoint
	if err = json.Unmarshal([]byte(regStr), &endpoint); err != nil {
		return nil, nil, err
	}

	return &endpoint, session, nil
}

// 生成EMAIL
func smailGenerate(suffix, proxy string, endpoint *smailEndpoint, session *requests.Session) (*requests.Session, error) {
	key, session, err := smailKey("", suffix, -1, proxy, session)
	if err != nil {
		return nil, err
	}

	params := map[string]any{
		"key":          key,
		"rapidapi-key": endpoint.Key,
		"domain":       suffix,
		"username":     "random",
		"server":       "server-1",
		"type":         "alias",
	}
	t := smailMatchType(suffix)
	response, session, err := newRequest(5*time.Second, proxy, http.MethodGet, endpoint.Endpoint+"/email/"+t+"/get", params, session)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Text)
	}
	obj, err := response.Json()
	if err != nil {
		return nil, errors.New("[json parsing error] 获取新Email失败")
	}
	if obj["code"].(float64) != 200 {
		return nil, errors.New(obj["msg"].(string))
	}

	items := obj["items"].(map[string]any)
	endpoint.Email = items["email"].(string)
	endpoint.Timestamp = items["timestamp"].(float64)
	return session, nil
}

func smailReceive(endpoint smailEndpoint, proxy string, session *requests.Session) (string, error) {
	var key string
	var response *models.Response
	key, session, err := smailKey(endpoint.Email, "", endpoint.Timestamp, proxy, session)
	if err != nil {
		return "", err
	}

	params := map[string]any{
		"key":          key,
		"rapidapi-key": endpoint.Key,
		"email":        endpoint.Email,
		"timestamp":    endpoint.Timestamp,
	}

	t := smailMatchType(endpoint.Email)
	response, _, err = newRequest(5*time.Second, proxy, http.MethodGet, endpoint.Endpoint+"/email/"+t+"/check", params, session)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New("[FetchError]: " + response.Text)
	}

	var obj map[string]interface{}
	obj, err = response.Json()
	if err != nil {
		return "", err
	}

	if obj["code"].(float64) != 200 {
		return "", errors.New("email_hme_check::" + obj["msg"].(string))
	}

	if emailSlice, ok := obj["items"].([]any); ok && len(emailSlice) > 0 {
		subject := emailSlice[0].(map[string]any)
		if strings.TrimSpace(subject["textFrom"].(string)) == "Anthropic" {
			sp := strings.Split(subject["textSubject"].(string), " ")
			code := sp[len(sp)-1]
			if code == "" {
				return "", errors.New("email_hme_check::解析验证码失败")
			}
			return code, nil
		}
	}

	return "", errors.New("email_hme_check::接收验证码失败")
}

func smailKey(email, suffix string, timestamp float64, proxy string, session *requests.Session) (string, *requests.Session, error) {
	var params map[string]any
	if email == "" {
		params = map[string]any{"domain": suffix, "username": "random", "server": "server-1", "type": "alias"}
	} else {
		params = map[string]any{"email": email, "timestamp": timestamp}
	}
	response, session, err := newRequest(15*time.Second, proxy, http.MethodPost, string(ED)+"app/key", params, session)
	if err != nil {
		return "", session, err
	}
	if response.StatusCode != 200 {
		return "", session, errors.New(response.Text)
	}
	obj, err := response.Json()
	if err != nil {
		return "", session, err
	}
	if obj["code"].(float64) != 200 {
		return "", session, errors.New(obj["msg"].(string))
	}
	return obj["items"].(string), session, nil
}

func smailMatchType(email string) string {
	r := "hme"
	if strings.HasSuffix(email, string(ES[0])) {
		r = "gm"
	}
	if strings.HasSuffix(email, string(ES[1])) {
		r = "ot/v2"
	}
	//if strings.HasSuffix(email, string(ES[2])) {
	//	r = "hme"
	//}
	return r
}
