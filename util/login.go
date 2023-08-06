package util

import (
	"errors"
	"github.com/bincooo/requests"
	"github.com/bincooo/requests/models"
	"github.com/bincooo/requests/url"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	rk = ""
	rt = ""
)

const (
	WebClaude2BU = "https://claude.ai/"
	JA3          = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"
	UA           = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
)

type Kv = map[string]string

func init() {
	_ = godotenv.Load()
	rk = loadEnvVar("RECAPTCHA_KEY", "")
	rt = loadEnvVar("RECAPTCHA_TOKEN", "")
}

func loadEnvVar(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func Login(proxy string) (string, error) {
	// validate
	if rk == "" || rt == "" {
		return "", errors.New("请在同级目录下的 .env 文件内配置 RECAPTCHA_KEY、RECAPTCHA_TOKEN 变量")
	}

	email, session, err := partOne()
	if err != nil {
		return "", err
	}
	return partTwo(proxy, email, session)
}

// create email
func partOne() (string, *requests.Session, error) {
	response, session, err := newRequest(15*time.Second, "", http.MethodGet, "https://www.guerrillamail.com/inbox", nil, nil)
	if err != nil {
		return "", nil, err
	}
	if response.StatusCode != 200 {
		return "", nil, errors.New("create_email Error: " + strconv.Itoa(response.StatusCode) + " Text=" + response.Text)
	}
	compileRegex := regexp.MustCompile(`Email:\s\S+@sharklasers.com`)
	matchSlice := compileRegex.FindStringSubmatch(response.Text)
	if len(matchSlice) == 0 {
		return "", nil, errors.New("create_email error")
	}
	return matchSlice[0][7:], session, nil
}

// send_code
func partTwo(proxy string, email string, session *requests.Session) (string, error) {
	response, _, err := newRequest(15*time.Second, proxy, http.MethodPost, WebClaude2BU+"api/auth/send_code", map[string]any{
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

	code, err := partThree(email, session)
	if err != nil {
		return "", err
	}

	return partFour(code, email, proxy)
}

// 注册成功，返回token
func partFour(code string, email string, proxy string) (string, error) {
	response, _, err := newRequest(15*time.Second, proxy, http.MethodPost, WebClaude2BU+"api/auth/verify_code", map[string]any{
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

func partThree(email string, session *requests.Session) (string, error) {
	slice := strings.Split(email, "@")
	cnt := 10
	for {
		cnt--
		if cnt < 0 {
			return "", errors.New("接收邮件失败")
		}
		response, _, err := newRequest(15*time.Second, "",
			http.MethodGet,
			"https://www.guerrillamail.com/ajax.php?f=get_email_list&offset=0&site="+slice[1]+"&in="+slice[0],
			nil,
			session)
		if err != nil {
			return "", err
		}

		json, err := response.Json()
		if err != nil {
			return "", err
		}

		if emailSlice, ok := json["list"].([]any); ok && len(emailSlice) > 1 {
			subject := emailSlice[0].(map[string]any)
			if subject["mail_from"] == "support@mail.anthropic.com" {
				split := strings.Split(subject["mail_subject"].(string), " ")
				return split[len(split)-1], nil
			}
		}
		time.Sleep(3 * time.Second)
	}
}

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

	uHeaders := url.NewHeaders()
	uHeaders.Set("User-Agent", UA)
	req.Headers = uHeaders

	req.Ja3 = JA3
	if session == nil {
		session = requests.NewSession()
	}
	response, err := session.Request(method, route, req, false)
	return response, session, err
}
