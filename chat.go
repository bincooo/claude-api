package claude

import (
	"errors"
	"fmt"
	"github.com/bincooo/claude-api/internal"
	"github.com/bincooo/claude-api/types"
	"github.com/bincooo/claude-api/vars"
	"strings"
)

func NewDefaultOptions(token string, model string) types.Options {
	options := types.Options{
		Retry: 2,
		Model: model,
	}

	switch model {
	case vars.Model4Slack:
		options.Headers = map[string]string{
			"Authorization": "Bearer " + token,
		}
	case vars.Model4WebClaude2:
		if token != "" && !strings.Contains(token, "sessionKey=") {
			token = "sessionKey=" + token
		}
		options.Headers = map[string]string{
			"cookie": token,
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
		return nil, errors.New(fmt.Sprintf("unknown model: `%v`", opt.Model))
	}
}
