package util

import (
	"encoding/json"
	"errors"
	"github.com/bincooo/requests"
	"github.com/bincooo/requests/models"
	"github.com/bincooo/requests/url"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	JA3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"

	rk  = ""
	rt  = ""
	rev = ""

	ED = []byte{104, 116, 116, 112, 115, 58, 47, 47, 115, 109, 97, 105, 108, 112, 114, 111, 46, 99, 111, 109, 47}
	ES = [][]byte{
		{103, 109, 97, 105, 108, 46, 99, 111, 109},
		{111, 117, 116, 108, 111, 111, 107, 46, 99, 111, 109},
		{105, 99, 108, 111, 117, 100, 46, 99, 111, 109},
	}
	//ED = []byte{104, 116, 116, 112, 115, 58, 47, 47, 119, 119, 119, 46, 103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 99, 111, 109, 47}
	//ES = [][]byte{
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 105, 110, 102, 111},
	//	{103, 114, 114, 46, 108, 97},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 98, 105, 122},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 99, 111, 109},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 100, 101},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 110, 101, 116},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 46, 111, 114, 103},
	//	{103, 117, 101, 114, 114, 105, 108, 108, 97, 109, 97, 105, 108, 98, 108, 111, 99, 107, 46, 99, 111, 109},
	//	{112, 111, 107, 101, 109, 97, 105, 108, 46, 110, 101, 116},
	//	{115, 112, 97, 109, 52, 46, 109, 101},
	//}
	//ED = []byte{104, 116, 116, 112, 115, 58, 47, 47, 119, 119, 119, 46, 108, 105, 110, 115, 104, 105, 121, 111, 117, 120, 105, 97, 110, 103, 46, 110, 101, 116, 47}
	//ES = [][]byte{
	//	{108, 105, 110, 115, 104, 105, 121, 111, 117, 120, 105, 97, 110, 103, 46, 110, 101, 116},
	//	{101, 117, 114, 45, 114, 97, 116, 101, 46, 99, 111, 109},
	//	{100, 101, 101, 112, 121, 105, 110, 99, 46, 99, 111, 109},
	//	{98, 101, 115, 116, 116, 101, 109, 112, 109, 97, 105, 108, 46, 99, 111, 109},
	//	{53, 108, 101, 116, 116, 101, 114, 119, 111, 114, 100, 115, 102, 105, 110, 100, 101, 114, 46, 99, 111, 109},
	//	{99, 101, 108, 101, 98, 114, 105, 116, 121, 100, 101, 116, 97, 105, 108, 101, 100, 46, 99, 111, 109},
	//	{99, 111, 109, 112, 97, 114, 105, 115, 105, 111, 110, 115, 46, 110, 101, 116},
	//	{114, 97, 110, 100, 111, 109, 112, 105, 99, 107, 101, 114, 115, 46, 99, 111, 109},
	//	{98, 101, 115, 116, 119, 104, 101, 101, 108, 115, 112, 105, 110, 110, 101, 114, 46, 99, 111, 109},
	//	{106, 117, 115, 116, 100, 101, 102, 105, 110, 105, 116, 105, 111, 110, 46, 99, 111, 109},
	//}
)

const (
	WebClaude2BU = "https://claude.ai/api"
	UA           = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
	InnerRk      = "6LcdsFgmAAAAAMfrnC1hEdmeRQRXCjpy8qT_kvfy"
	InnerRt      = "03AAYGu2Ru5V3hpGin5CiSBezXZ5xIKLHhaU7tJ7n2JIbqwnt9WIFoI9PEB4UHEfdHDXsGmfv7H1dRn8jzguQp8KJgLfMrz6jK2pKwt0G8SApU9zJ-LOul39kseZwtONBr-N1EkFQsF7NrPrQUiFtdrJ1g0ZvDZKbhlo6iQCbu2laB8ieumQP3h-PxCbREOt2dzw7NJFKrjW8R8sNJdi6tuKMN7q89ant-llQgA4ZuvzU8Qkf52nkLqMkdpfNpE1n7pkXng8EDovH2i1pXCx-BwkAR-3ihiPqK-6npFA0L3VuzM1MPMlhuDJloLSDj6o8QZXqeOOIGrW-OHoXbDtG7hzUd8rN_m3eeslu0Eis9jZWZi41Y-9-gU6KPOEuy8SGA_HmccK5ziEYJXqcA5KwsoKJ50ydblTfm259S614m4Bn1OmlTucFYOK9Kk5Km1-TO6OGLy9ZEqFSeR2tiKWaObxvcE56HV-qNONc2bIfBdjAtFgFOfbqicSfiwY7FjZO11VYSXytMDPxGLSRQ9vqNW19pVN8ew3khdajb8XeCsuuKGxdRPzBpyd0VnZ54736EweKfT83fzCDMc_bcsTQVVSUAp6XIhQWRHpZbzV-OoNNIJO5mWh6xH8Vq-Pc9r2kHBcikgVavaW5Oawxo99HAsTdbOM4G2pDxRaOhySJyoTT5yOlxz9fxIiVxH76BMBc1ImRRSDvJzm7oMWpmk1twg0xzVdw4W1aq7i0ErI_Cd_vd4mcfpZodbSP9NJUskdBw_YH0d1Bqe_ApzhqUJU9TiJE7GMU1_gqJGaxQRdL1SwDaJZ-yoL3xe8B47NX2GnBdBT68LXl7zmkZETsagtusTwQVCNiimvUnVhqSdpSIu-3CXOlCPIPMIJyKUvYj7RRRmNFTlV62MZghyirmOomTHEs9h_nVkb0Tcm_JU9F0iBXju3OidZXEwXKRxuXR2M13qlEuSHRxM0jl6V-wuCIb2ImunOpQZ7DPQkjO9Pxkpcxs0YnGhbIW_EteU8tLYo8xCznonDk4wFhs6SafzQT3ApMLnMxOcZyKcxAUj3Zj-6Bwq03RmzImTOackPYR3TWcANNXfdEIWgNvT-SKV4e04d1HjgCF3YRXCupT83QTQOjhYUUsEhD-oA_W42VIEWI51SzqKKAujwZ1hlIZjZi5QPppCVYjpBLPuFjhmxiRpJ8irdq4XQN7Y4CDMr_6A8GX7epLxUR0z9x8DaTQVE3NdKL7MfVin-CdJyt2EiGJz9QroExEm7ohjd4LNzLleHd_1s6FmwZLWl8ucSFx_SZwF9_49zT5v_tzXM5EAZRHXIuCIrMmsM4ShmS0_dPn6VYTe7A2L7EfCmKbcK5wNdc2xQaYPnXNnt5e7Lb6UCf6B0IwVUt20CTkII5wUHafxomGUtQYyvEWQZmt6cA_mhxcI2Fl6e7Rdv3HkN1blEnEnOw_nX2iwsuRJI9Kmia1JTOEtoQrWpdhM5Ogh6PtcMcQtUGLwDPl0av-gL7IfQkkFTwzfJlMwC1B2ogNsxo_q83w8jT4hLbusYBgPOGKQbsuPY_quZW-sCOS_BFt8W5xtS5wZbWtuzTyS8Dca7uSsW6_Y6Ak2VehP5ZlIwsh6Ic1D_L17aBQVi6HWBIPvI8_o8L1_vc_tgcxv9O5HKX3hS5bxCLzlXxVlcxleBUhrdDphKfSErVI0hEEp4eh1RAksSWQKQvfFn68WLGuqn4AYUjW9KPrBxDlN7sl14NFeq8x7QfbnNVrXmmN4g"
)

type Kv = map[string]string

func init() {
	_ = godotenv.Load()
	JA3 = LoadEnvVar("JA3", JA3)
	rk = LoadEnvVar("RECAPTCHA_KEY", "")
	rt = LoadEnvVar("RECAPTCHA_TOKEN", "")
	rev = LoadEnvVar("REV", "")
}

func LoadEnvVar(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func Login(proxy string) (string, string, error) {
	return LoginFor("", "", proxy)
}

func LoginFor(baseURL, suffix, proxy string) (string, string, error) {
	// validate
	//if rk == "" || rt == "" {
	//	logrus.Warning("你没有提供`RECAPTCHA_KEY`、`RECAPTCHA_TOKEN`，使用内置参数；如若无法生成请在同级目录下的 .env 文件内配置 RECAPTCHA_KEY、RECAPTCHA_TOKEN 变量")
	//	rk = InnerRk
	//	rt = InnerRt
	//}

	if baseURL == "" {
		baseURL = WebClaude2BU
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	retry := 3
	var err error
	var email string

	for {
		retry--
		if retry < 0 {
			if err != nil {
				return email, "", err
			}
			return email, "", errors.New("获取SessionKey失败")
		}
		//em, session, e := partOne(suffix, proxy)
		endpoint, em, session, e := partOne(suffix, proxy)
		if e != nil {
			err = e
			continue
		}
		email = em
		token, e := partTwo(endpoint, baseURL, proxy, email, session)
		if e != nil {
			err = e
			continue
		}
		return email, token, e
	}
}

// create email
func partOne(suffix, proxy string) (string, string, *requests.Session, error) {
	if suffix == "" {
		suffix = string(ES[rand.Intn(len(ES))])
	}
	params := map[string]any{
		"id": RandHexString(20),
	}
	response, session, err := newRequest(5*time.Second, proxy, http.MethodGet, string(ED)+"js/chunks/smailpro_v2_email.js", params, nil)
	if err != nil {
		return "", "", nil, err
	}
	if response.StatusCode != 200 {
		return "", "", nil, errors.New("smailpro_v2_email.js::405 Not Allowed")
	}
	compileRegex := regexp.MustCompile(`{\s*rapidapi_endpoint:[^}]+}[\s\n]*}`)
	matchStr := compileRegex.FindString(response.Text)
	if len(matchStr) == 0 {
		return "", "", nil, errors.New("create_email error::not match rapidapi_endpoint")
	}

	reg := regexp.MustCompile(`([^",{][^(http)][a-z_A-Z]\w*[^"]):`)
	regStr := reg.ReplaceAllString(matchStr, `"$1":`)
	var endpoint map[string]any
	if err = json.Unmarshal([]byte(regStr), &endpoint); err != nil {
		return "", "", nil, err
	}

	rapidapiKey := endpoint["rapidapi_key"].(string)
	rapidapiEndpoint := endpoint["rapidapi_endpoint"].(string)

	key, session, err := getKey("", suffix, proxy, session)
	if err != nil {
		return "", "", nil, err
	}

	params = map[string]any{
		"key":          key,
		"rapidapi-key": rapidapiKey,
		"domain":       suffix,
		"username":     "random",
		"server":       "server-1",
		"type":         "alias",
	}
	r := gr(suffix)
	response, session, err = newRequest(5*time.Second, proxy, http.MethodGet, rapidapiEndpoint+"/email/"+r+"/get", params, session)
	if err != nil {
		return "", "", nil, err
	}
	if response.StatusCode != 200 {
		return "", "", nil, errors.New(response.Text)
	}
	obj, err := response.Json()
	if err != nil {
		return "", "", nil, errors.New("json parsing error")
	}
	if obj["code"].(float64) != 200 {
		return "", "", nil, errors.New(obj["msg"].(string))
	}
	items := obj["items"].(map[string]any)
	return rapidapiEndpoint + "###" + rapidapiKey, items["email"].(string), session, nil
}

func getKey(email, suffix, proxy string, session *requests.Session) (string, *requests.Session, error) {
	var params map[string]any
	if email == "" {
		params = map[string]any{"domain": suffix, "username": "random", "server": "server-1", "type": "alias"}
	} else {
		params = map[string]any{"email": email, "timestamp": time.Now().Unix()}
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

//func partOne(suffix, proxy string) (string, *requests.Session, error) {
//	response, session, err := newRequest(5*time.Second, proxy, http.MethodGet, string(ED)+"api/v1/mailbox/keepalive", nil, nil)
//	if err != nil {
//		return "", nil, err
//	}
//	if response.StatusCode != 200 {
//		return "", nil, errors.New("create_email Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
//	}
//	obj, err := response.Json()
//	if err != nil {
//		return "", nil, err
//	}
//
//	if suffix == "" {
//		suffix = string(ES[rand.Intn(len(ES))])
//	}
//	email := obj["mailbox"].(string) + "@" + suffix
//	return email, session, nil
//}

//func partOne(suffix, proxy string) (string, *requests.Session, error) {
//	response, session, err := newRequest(15*time.Second, proxy, http.MethodGet, string(ED)+"inbox", nil, nil)
//	if err != nil {
//		return "", nil, err
//	}
//	if response.StatusCode != 200 {
//		return "", nil, errors.New("create_email Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
//	}
//	compileRegex := regexp.MustCompile(`Email:\s\S+@sharklasers.com`)
//	matchSlice := compileRegex.FindStringSubmatch(response.Text)
//	if len(matchSlice) == 0 {
//		return "", nil, errors.New("create_email error")
//	}
//
//	if suffix == "" {
//		suffix = string(ES[rand.Intn(len(ES))])
//	}
//
//	email := strings.Replace(matchSlice[0][7:], "@sharklasers.com", "@"+suffix, -1)
//	return email, session, nil
//}

// send_code
func partTwo(endpoint, baseURL, proxy string, email string, session *requests.Session) (string, error) {
	if rev != "" {
		response, _, err := newRequest(30*time.Second, "", http.MethodPost, rev+"/send_code", map[string]any{
			"email_address": email,
		}, nil)
		if err != nil {
			return "", err
		}

		if response.StatusCode != 200 {
			return "", errors.New("send_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}

		result, err := response.Json()
		if err != nil {
			return "", errors.New("send_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}
		if result["code"].(float64) != 200 {
			return "", errors.New(result["msg"].(string))
		}

	} else {
		response, _, err := newRequest(5*time.Second, proxy, http.MethodPost, baseURL+"auth/send_code", map[string]any{
			"email_address":      email,
			"recaptcha_site_key": rk,
			"recaptcha_token":    rt,
		}, nil)
		if err != nil {
			return "", err
		}

		if response.StatusCode != 200 {
			return "", errors.New("send_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}

		if response.Text != `{"success":true}` {
			return "", errors.New("send_code Error: " + response.Text)
		}
	}

	//code, err := partThree(email, proxy, session)
	code, err := partThree(endpoint, email, proxy, session)
	if err != nil {
		if rev != "" && !strings.Contains(rev, "claudeai.ai") {
			// 接收失败清理
			_, _, _ = newRequest(30*time.Second, "", http.MethodGet, rev+"/delete/"+email, nil, nil)
		}
		return "", err
	}
	return partFour(baseURL, code, email, proxy)
}

// 注册成功，返回token
func partFour(baseURL, code string, email string, proxy string) (string, error) {
	if rev != "" {
		response, _, err := newRequest(30*time.Second, "", http.MethodPost, rev+"/verify_code", map[string]any{
			"email_address": email,
			"verify_code":   code,
		}, nil)
		if err != nil {
			return "", err
		}

		if response.StatusCode != 200 {
			return "", errors.New("verify_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}

		sc := response.Headers.Get("Set-Cookie")
		if strings.HasPrefix(sc, "sessionKey") {
			slice := strings.Split(sc, ";")
			return slice[0][11:], nil
		}

		result, err := response.Json()
		if err != nil {
			return "", errors.New("verify_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}
		if result["code"].(float64) != 200 {
			return "", errors.New(result["message"].(string))
		}

		return result["msg"].(string), nil

	} else {
		response, _, err := newRequest(10*time.Second, proxy, http.MethodPost, baseURL+"auth/verify_code", map[string]any{
			"code":               code,
			"email_address":      email,
			"recaptcha_site_key": rk,
			"recaptcha_token":    rt,
		}, nil)
		if err != nil {
			return "", err
		}

		if response.StatusCode != 200 {
			return "", errors.New("verify_code Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
		}

		if response.Text != `{"success":true}` {
			return "", errors.New("verify_code Error: " + response.Text)
		}

		sc := response.Headers.Get("Set-Cookie")
		if !strings.HasPrefix(sc, "sessionKey") {
			return "", errors.New("resolve Set-Cookie error")
		}
		slice := strings.Split(sc, ";")
		return slice[0][11:], nil
	}
}

// 接收验证码
func partThree(endpoint, email, proxy string, session *requests.Session) (string, error) {
	split := strings.Split(endpoint, "###")
	endpoint = split[0]

	key, session, err := getKey(email, "", proxy, session)
	if err != nil {
		return "", err
	}

	params := map[string]any{
		"key":          key,
		"rapidapi-key": split[1],
		"email":        email,
		"timestamp":    time.Now().Unix(),
	}

	r := gr(email)

	cnt := 18
	for {
		cnt--
		if cnt < 0 {
			return "", errors.New("接收邮件失败")
		}
		response, _, err := newRequest(5*time.Second, proxy, http.MethodGet, endpoint+"/email/"+r+"/check", params, session)
		if err != nil {
			return "", err
		}

		if response.StatusCode != 200 {
			return "", errors.New("[FetchError]: " + response.Text)
		}

		obj, err := response.Json()
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
					continue
				}
				return code, nil
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func gr(email string) string {
	r := "hme"
	if strings.HasSuffix(email, string(ES[0])) {
		r = "gm"
	}
	if strings.HasSuffix(email, string(ES[1])) {
		r = "ot"
	}
	if strings.HasSuffix(email, string(ES[2])) {
		r = "hme"
	}
	return r
}

//func partThree(email, proxy string, session *requests.Session) (string, error) {
//	slice := strings.Split(email, "@")
//	cnt := 10
//	for {
//		cnt--
//		if cnt < 0 {
//			return "", errors.New("接收邮件失败")
//		}
//		response, _, err := newRequest(10*time.Second, proxy,
//			http.MethodGet,
//			string(ED)+"api/v1/mailbox/"+slice[0],
//			nil,
//			session)
//		if err != nil {
//			return "", err
//		}
//
//		if response.StatusCode != 200 {
//			return "", errors.New("[FetchError]: " + response.Text)
//		}
//
//		var emailSlice []map[string]any
//		if err = json.Unmarshal([]byte(response.Text), &emailSlice); err != nil {
//			return "", err
//		}
//
//		if len(emailSlice) > 0 {
//			subject := emailSlice[0]
//			if subject["from"] == "Anthropic" {
//				split := strings.Split(subject["subject"].(string), " ")
//				return split[len(split)-1], nil
//			}
//		}
//		time.Sleep(3 * time.Second)
//	}
//}

//func partThree(email, proxy string, session *requests.Session) (string, error) {
//	slice := strings.Split(email, "@")
//	cnt := 10
//	for {
//		cnt--
//		if cnt < 0 {
//			return "", errors.New("接收邮件失败")
//		}
//		response, _, err := newRequest(15*time.Second, proxy,
//			http.MethodGet,
//			string(ED)+"ajax.php?f=get_email_list&offset=0&site=guerrillamail.com&in="+slice[0],
//			nil,
//			session)
//		if err != nil {
//			return "", err
//		}
//
//		if response.StatusCode != 200 {
//			return "", errors.New("[FetchError]: " + response.Text)
//		}
//
//		json, err := response.Json()
//		if err != nil {
//			return "", err
//		}
//
//		if emailSlice, ok := json["list"].([]any); ok && len(emailSlice) > 1 {
//			subject := emailSlice[0].(map[string]any)
//			if subject["mail_from"] == "support@mail.anthropic.com" {
//				split := strings.Split(subject["mail_subject"].(string), " ")
//				return split[len(split)-1], nil
//			}
//		}
//		time.Sleep(3 * time.Second)
//	}
//}

func newRequest(timeout time.Duration, proxy string, method string, route string, params map[string]any, session *requests.Session) (*models.Response, *requests.Session, error) {
	if method == http.MethodGet {
		var search []string
		for key, value := range params {
			if v, ok := value.(string); ok {
				search = append(search, key+"="+v)
			}
		}

		if len(search) > 0 {
			route += "?" + strings.Join(search, "&")
		}

		params = nil
	}

	req := url.NewRequest()
	req.Timeout = timeout
	if method != http.MethodGet && params != nil {
		req.Json = params
	}

	if proxy != "" {
		req.Proxies = proxy
	}

	parse, err := url.Parse(route)
	if err != nil {
		return nil, nil, err
	}

	uHeaders := url.NewHeaders()
	uHeaders.Set("User-Agent", UA)
	uHeaders.Set("referer", parse.Scheme+"://"+parse.Host)
	req.Headers = uHeaders

	req.Ja3 = JA3
	if session == nil {
		session = requests.NewSession()
	}
	response, err := session.Request(method, route, req, false)
	return response, session, err
}

func RandHexString(length int) string {
	hexStr := "01a23b45c67d89e"
	b := make([]byte, length)
	for i := range b {
		b[i] = hexStr[rand.Intn(len(hexStr))]
	}
	return string(b)
}

// 缓存Key
func cacheKey(key string) {
	// 文件不存在...   就创建吧
	if _, err := os.Lstat(".env"); os.IsNotExist(err) {
		if _, e := os.Create(".env"); e != nil {
			logrus.Error(e)
			return
		}
	}

	bytes, err := os.ReadFile(".env")
	if err != nil {
		logrus.Error(err)
	}
	tmp := string(bytes)
	compileRegex := regexp.MustCompile(`(\n|^)CACHE_SMAIL_PRO\s*=[^\n]*`)
	matchSlice := compileRegex.FindStringSubmatch(tmp)
	if len(matchSlice) > 0 {
		str := matchSlice[0]
		if strings.HasPrefix(str, "\n") {
			str = str[1:]
		}
		tmp = strings.Replace(tmp, str, "CACHE_SMAIL_PRO=\""+key+"\"", -1)
	} else {
		delimiter := ""
		if len(tmp) > 0 && !strings.HasSuffix(tmp, "\n") {
			delimiter = "\n"
		}
		tmp += delimiter + "CACHE_SMAIL_PRO=\"" + key + "\""
	}
	err = os.WriteFile(".env", []byte(tmp), 0664)
	if err != nil {
		logrus.Error(err)
	}
}
