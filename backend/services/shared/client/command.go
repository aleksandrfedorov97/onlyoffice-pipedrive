package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var ErrCommandServiceError = errors.New("got a command service error 1 status")

type CommandClient struct {
	client     *resty.Client
	jwtManager crypto.JwtManager
}

func NewCommandClient(jwtManager crypto.JwtManager) CommandClient {
	otelClient := otelhttp.DefaultClient
	otelClient.Transport = otelhttp.NewTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})
	return CommandClient{
		client: resty.NewWithClient(otelClient).
			SetRetryCount(0).
			SetRetryWaitTime(120 * time.Millisecond).
			SetRetryMaxWaitTime(900 * time.Millisecond).
			SetLogger(log.NewEmptyLogger()).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests
			}),
		jwtManager: jwtManager,
	}
}

func (p *CommandClient) License(ctx context.Context, url, secret string) error {
	var resp response.BaseCommandResponse

	token, err := p.jwtManager.Sign(secret, request.BaseCommandRequest{
		C: "version",
	})

	if err != nil {
		return err
	}

	res, err := p.client.R().
		SetContext(ctx).
		SetBody(request.TokenCommandRequest{
			Token: token,
		}).
		SetResult(&resp).
		Post(fmt.Sprintf("%scoauthoring/CommandService.ashx", url))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 300 || resp.Error != 0 {
		return ErrCommandServiceError
	}

	return nil
}
