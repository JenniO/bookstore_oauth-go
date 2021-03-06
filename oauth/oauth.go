package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JenniO/bookstore_utils-go/logger"
	"github.com/JenniO/bookstore_utils-go/rest_errors"
	"github.com/federicoleon/golang-restclient/rest"
	"net/http"
	"strconv"
	"time"
)

const (
	headerXPublic    = "X-Public"
	headerXClientId  = "X-Client-Id"
	headerXCallerId  = "X-Caller-Id"
	paramAccessToken = "access_token"
)

var (
	oauthRestClient = rest.RequestBuilder{
		BaseURL: "http://localhost:8084",
		Timeout: 200 * time.Millisecond,
	}
)

type accessToken struct {
	Id       string `json:"id"`
	UserId   int64  `json:"user_id"`
	ClientId int64  `json:"client_id"`
}

func IsPublic(request *http.Request) bool {
	if request == nil {
		return true
	}
	return request.Header.Get(headerXPublic) == "true"
}

func GetCallerId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	callerId, err := strconv.ParseInt(request.Header.Get(headerXCallerId), 10, 64)
	if err != nil {
		return 0
	}
	return callerId
}

func GetClientId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	clientId, err := strconv.ParseInt(request.Header.Get(headerXClientId), 10, 64)
	if err != nil {
		return 0
	}
	return clientId
}

func AuthenticateRequest(request *http.Request) rest_errors.RestErr {
	if request == nil {
		return nil
	}

	cleanRequest(request)

	accessTokenId := request.URL.Query().Get(paramAccessToken)
	if accessTokenId == "" {
		return nil
	}

	at, err := getAccessToken(accessTokenId)
	if err != nil {
		if err.Status() == http.StatusNotFound {
			return nil
		}
		return err
	}

	request.Header.Add(headerXClientId, strconv.FormatInt(at.ClientId, 10))
	request.Header.Add(headerXCallerId, strconv.FormatInt(at.UserId, 10))
	return nil
}

func cleanRequest(request *http.Request) {
	if request == nil {
		return
	}
	request.Header.Del(headerXClientId)
	request.Header.Del(headerXCallerId)
}

func getAccessToken(accessTokenId string) (*accessToken, rest_errors.RestErr) {
	path := fmt.Sprintf("/oauth/access_token/%s", accessTokenId)
	logger.Info(path)
	response := oauthRestClient.Get(path)
	if response == nil || response.Response == nil {
		logger.Info("invalid rest client response, when trying to get access token")
		return nil, rest_errors.NewInternalServerError("invalid rest client response, when trying to get access token", errors.New("oauth error"))
	}

	if response.StatusCode > 299 {
		logger.Info(response.String())
		var restErr rest_errors.RestErr
		var err error
		if restErr, err = rest_errors.NewRestErrorFromBytes(response.Bytes()); err != nil {
			logger.Info("invalid error interface when trying to get access token")
			return nil, rest_errors.NewInternalServerError("invalid error interface when trying to get access token", err)
		}

		return nil, restErr
	}

	var at accessToken
	if err := json.Unmarshal(response.Bytes(), &at); err != nil {
		logger.Info(response.String())
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal access token response", err)
	}
	return &at, nil
}
