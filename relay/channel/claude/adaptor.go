package claude

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"strings"
)

const (
	RequestModeCompletion = 1
	RequestModeMessage    = 2
)

type Adaptor struct {
	RequestMode int
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	if strings.HasPrefix(info.UpstreamModelName, "claude-3") {
		a.RequestMode = RequestModeMessage
	} else {
		a.RequestMode = RequestModeCompletion
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if a.RequestMode == RequestModeMessage {
		return fmt.Sprintf("%s/v1/messages", info.BaseUrl), nil
	} else {
		return fmt.Sprintf("%s/v1/complete", info.BaseUrl), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Header.Set("x-api-key", info.ApiKey)

	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Header.Set("anthropic-version", anthropicVersion)

	anthropicBeta := c.Request.Header.Get("anthropic-beta")
	if "" != anthropicBeta {
		req.Header.Set("anthropic-beta", anthropicBeta)
	}

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if a.RequestMode == RequestModeCompletion {
		return RequestOpenAI2ClaudeComplete(c, *request), nil
	} else {
		return RequestOpenAI2ClaudeMessage(c, *request)
	}
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage *dto.Usage, err *dto.OpenAIErrorWithStatusCode) {
	if info.IsStream {
		err, usage = ClaudeStreamHandler(c, resp, info, a.RequestMode)
	} else {
		err, usage = ClaudeHandler(c, resp, a.RequestMode, info)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
