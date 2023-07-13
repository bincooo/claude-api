package claude

import (
	"errors"
	"fmt"
	"github.com/bincooo/claude-api/internal"
	"github.com/bincooo/claude-api/types"
	"github.com/bincooo/claude-api/vars"
)

func NewDefaultOptions(token string, botId string, model string) types.Options {
	options := types.Options{
		Retry: 2,
		BotId: botId,
		Model: model,
	}

	if model == vars.Model4Slack {
		options.Headers = map[string]string{
			"Authorization": "Bearer " + token,
		}
	}
	switch model {
	case vars.Model4Slack:
		options.Headers = map[string]string{
			"Authorization": "Bearer " + token,
		}
	case vars.Model4WebClaude2:
		options.Headers = map[string]string{
			"cookie": "sessionKey=" + token,
			//"authority":       "claude.ai",
			//"accept":          "text/event-stream",
			//"accept-language": "en,zh-CN;q=0.9,zh;q=0.8,en-GB;q=0.7,en-US;q=0.6",
			//"cache-control":   "no-cache",
			//"content-type":    "application/json",
			//"origin":          "https://claude.ai",
			//"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		}
	}

	return options
}

func New(opt types.Options) (types.Chat, error) {
	switch opt.Model {
	case vars.Model4Slack:
		return internal.NewSlack(opt), nil
	case vars.Model4WebClaude2:
		return internal.NewWebClaude2(opt), nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown model: `%v`", opt.Model))
	}
}
