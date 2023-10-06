package util

import (
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
	rev = ""

	ED = []byte{104, 116, 116, 112, 115, 58, 47, 47, 115, 109, 97, 105, 108, 112, 114, 111, 46, 99, 111, 109, 47}
	ES = [][]byte{
		{103, 109, 97, 105, 108, 46, 99, 111, 109},
		{111, 117, 116, 108, 111, 111, 107, 46, 99, 111, 109},
		//{105, 99, 108, 111, 117, 100, 46, 99, 111, 109},
	}
)

const (
	WebClaude2BU = "https://claude.ai/api"
	UA           = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
)

type Kv = map[string]string

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.Error(err)
	}
	JA3 = LoadEnvVar("JA3", JA3)
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

	if baseURL == "" {
		baseURL = WebClaude2BU
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	retry := 3
	var err error

	// ========================
	// 登陆失败重试3次，除非claude限流
	for {
		retry--
		if retry < 0 {
			if err != nil {
				return "", "", err
			}
			return "", "", errors.New("获取SessionKey失败")
		}
		endpoint, session, e := partOne(suffix, proxy)
		if e != nil {
			if strings.Contains(e.Error(), "Rate limited.") {
				return "", "", e
			}
			err = e
			continue
		}
		token, e := partTwo(endpoint, baseURL, proxy, session)
		if e != nil {
			if strings.Contains(e.Error(), "Rate limited.") {
				return endpoint.Email, "", e
			}
			err = e
			continue
		}
		return endpoint.Email, token, e
	}
}

// create email
func partOne(suffix, proxy string) (smailEndpoint, *requests.Session, error) {
	if suffix == "" {
		suffix = string(ES[rand.Intn(len(ES))])
	}

	// https://smailpro.com/advanced
	jsId, err := smailAdvancedAppJSID(proxy)
	if err != nil {
		return smailEndpoint{}, nil, err
	}

	v2Id, err := smailAppV2Id(jsId, proxy)
	if err != nil {
		return smailEndpoint{}, nil, err
	}

	endpoint, session, err := smailEndpointKey(v2Id, proxy)
	if err != nil {
		return smailEndpoint{}, nil, err
	}

	session, err = smailGenerate(suffix, proxy, endpoint, session)
	if err != nil {
		return smailEndpoint{}, nil, err
	}
	return *endpoint, session, nil
}

// send_code
func partTwo(endpoint smailEndpoint, baseURL, proxy string, session *requests.Session) (string, error) {
	if rev == "" {
		return "", errors.New("请先配置claude注册转发地址`REV`")
	}
	response, _, err := newRequest(30*time.Second, "", http.MethodPost, rev+"/send_code", map[string]any{
		"email_address": endpoint.Email,
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

	code, err := partThree(endpoint, proxy, session)
	if err != nil {
		if rev != "" {
			// 接收失败清理
			_, _, _ = newRequest(30*time.Second, "", http.MethodGet, rev+"/delete/"+endpoint.Email, nil, nil)
		}
		return "", err
	}
	return partFour(code, endpoint.Email)
}

// 接收验证码
func partThree(endpoint smailEndpoint, proxy string, session *requests.Session) (string, error) {
	cnt := 10
	var err error
	var code string

	// 轮询5s x 10次
	for {
		cnt--
		if cnt < 0 {
			if err != nil {
				return "", err
			}
			return "", errors.New("接收邮件失败")
		}
		code, err = smailReceive(endpoint, proxy, session)
		if err == nil && code != "" {
			return code, nil
		}
		time.Sleep(5 * time.Second)
	}
}

// 注册成功，返回token
func partFour(code string, email string) (string, error) {
	if rev == "" {
		return "", errors.New("请先配置claude注册转发地址`REV`")
	}
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

	if result["cookies"] != nil {
		sessionKey := result["cookies"].(map[string]any)["sessionKey"]
		if sessionKey != "" {
			return sessionKey.(string), nil
		}
	}
	return result["msg"].(string), nil
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
