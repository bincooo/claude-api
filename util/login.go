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
	rk = ""
	rt = ""

	emailSubs = []string{
		"guerrillamail.biz",
		"guerrillamail.de",
		"guerrillamail.net",
		"guerrillamail.org",
		"guerrillamail.info",
		"guerrillamailblock.com",
		"pokemail.net",
		"spam4.me",
		"grr.la",
	}
)

const (
	WebClaude2BU = "https://claude.ai/api"
	JA3          = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"
	UA           = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
	InnerRk      = "6LcdsFgmAAAAAMfrnC1hEdmeRQRXCjpy8qT_kvfy"
	InnerRt      = "03AAYGu2Ru5V3hpGin5CiSBezXZ5xIKLHhaU7tJ7n2JIbqwnt9WIFoI9PEB4UHEfdHDXsGmfv7H1dRn8jzguQp8KJgLfMrz6jK2pKwt0G8SApU9zJ-LOul39kseZwtONBr-N1EkFQsF7NrPrQUiFtdrJ1g0ZvDZKbhlo6iQCbu2laB8ieumQP3h-PxCbREOt2dzw7NJFKrjW8R8sNJdi6tuKMN7q89ant-llQgA4ZuvzU8Qkf52nkLqMkdpfNpE1n7pkXng8EDovH2i1pXCx-BwkAR-3ihiPqK-6npFA0L3VuzM1MPMlhuDJloLSDj6o8QZXqeOOIGrW-OHoXbDtG7hzUd8rN_m3eeslu0Eis9jZWZi41Y-9-gU6KPOEuy8SGA_HmccK5ziEYJXqcA5KwsoKJ50ydblTfm259S614m4Bn1OmlTucFYOK9Kk5Km1-TO6OGLy9ZEqFSeR2tiKWaObxvcE56HV-qNONc2bIfBdjAtFgFOfbqicSfiwY7FjZO11VYSXytMDPxGLSRQ9vqNW19pVN8ew3khdajb8XeCsuuKGxdRPzBpyd0VnZ54736EweKfT83fzCDMc_bcsTQVVSUAp6XIhQWRHpZbzV-OoNNIJO5mWh6xH8Vq-Pc9r2kHBcikgVavaW5Oawxo99HAsTdbOM4G2pDxRaOhySJyoTT5yOlxz9fxIiVxH76BMBc1ImRRSDvJzm7oMWpmk1twg0xzVdw4W1aq7i0ErI_Cd_vd4mcfpZodbSP9NJUskdBw_YH0d1Bqe_ApzhqUJU9TiJE7GMU1_gqJGaxQRdL1SwDaJZ-yoL3xe8B47NX2GnBdBT68LXl7zmkZETsagtusTwQVCNiimvUnVhqSdpSIu-3CXOlCPIPMIJyKUvYj7RRRmNFTlV62MZghyirmOomTHEs9h_nVkb0Tcm_JU9F0iBXju3OidZXEwXKRxuXR2M13qlEuSHRxM0jl6V-wuCIb2ImunOpQZ7DPQkjO9Pxkpcxs0YnGhbIW_EteU8tLYo8xCznonDk4wFhs6SafzQT3ApMLnMxOcZyKcxAUj3Zj-6Bwq03RmzImTOackPYR3TWcANNXfdEIWgNvT-SKV4e04d1HjgCF3YRXCupT83QTQOjhYUUsEhD-oA_W42VIEWI51SzqKKAujwZ1hlIZjZi5QPppCVYjpBLPuFjhmxiRpJ8irdq4XQN7Y4CDMr_6A8GX7epLxUR0z9x8DaTQVE3NdKL7MfVin-CdJyt2EiGJz9QroExEm7ohjd4LNzLleHd_1s6FmwZLWl8ucSFx_SZwF9_49zT5v_tzXM5EAZRHXIuCIrMmsM4ShmS0_dPn6VYTe7A2L7EfCmKbcK5wNdc2xQaYPnXNnt5e7Lb6UCf6B0IwVUt20CTkII5wUHafxomGUtQYyvEWQZmt6cA_mhxcI2Fl6e7Rdv3HkN1blEnEnOw_nX2iwsuRJI9Kmia1JTOEtoQrWpdhM5Ogh6PtcMcQtUGLwDPl0av-gL7IfQkkFTwzfJlMwC1B2ogNsxo_q83w8jT4hLbusYBgPOGKQbsuPY_quZW-sCOS_BFt8W5xtS5wZbWtuzTyS8Dca7uSsW6_Y6Ak2VehP5ZlIwsh6Ic1D_L17aBQVi6HWBIPvI8_o8L1_vc_tgcxv9O5HKX3hS5bxCLzlXxVlcxleBUhrdDphKfSErVI0hEEp4eh1RAksSWQKQvfFn68WLGuqn4AYUjW9KPrBxDlN7sl14NFeq8x7QfbnNVrXmmN4g"
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
	return LoginFor("", proxy)
}

func LoginFor(baseURL, proxy string) (string, error) {
	// validate
	if rk == "" || rt == "" {
		logrus.Warning("你没有提供`RECAPTCHA_KEY`、`RECAPTCHA_TOKEN`，使用内置参数；如若无法生成请在同级目录下的 .env 文件内配置 RECAPTCHA_KEY、RECAPTCHA_TOKEN 变量")
		// return "", errors.New("请在同级目录下的 .env 文件内配置 RECAPTCHA_KEY、RECAPTCHA_TOKEN 变量")
		rk = InnerRk
		rt = InnerRt
	}

	if baseURL == "" {
		baseURL = WebClaude2BU
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	email, session, err := partOne()
	if err != nil {
		return "", err
	}
	return partTwo(baseURL, proxy, email, session)
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

	email := strings.Replace(matchSlice[0][7:], "@sharklasers.com", "@"+emailSubs[rand.Intn(len(emailSubs))], -1)
	return email, session, nil
}

// send_code
func partTwo(baseURL, proxy string, email string, session *requests.Session) (string, error) {
	response, _, err := newRequest(15*time.Second, proxy, http.MethodPost, baseURL+"auth/send_code", map[string]any{
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

	return partFour(baseURL, code, email, proxy)
}

// 注册成功，返回token
func partFour(baseURL, code string, email string, proxy string) (string, error) {
	response, _, err := newRequest(15*time.Second, proxy, http.MethodPost, baseURL+"auth/verify_code", map[string]any{
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
