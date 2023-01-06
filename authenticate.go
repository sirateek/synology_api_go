package synologyapi

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/google/go-querystring/query"
	"github.com/sirateek/synologyapi/models"
)

type authenticate struct {
	BaseApi
	Api     string
	cgiPath string
}

type Authenticate interface {
	Login(credential models.ApiCredential) (sid string, err error)
}

func NewAuthenticate(baseApi BaseApi) Authenticate {
	return &authenticate{
		BaseApi: baseApi,
		Api:     "SYNO.API.Auth",
		cgiPath: "auth.cgi",
	}
}

func (a *authenticate) Login(credential models.ApiCredential) (string, error) {
	// Prepare the query params.
	value, err := query.Values(credential)
	if err != nil {
		return "", err
	}
	value.Add("api", a.Api)
	value.Add("method", "login")
	value.Add("version", "3")

	// Inject the param into the request.
	req, err := a.GetNewHttpRequest(GET, a.cgiPath)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = value.Encode()

	response, err := a.GetHttpClient().Do(req)
	if err != nil {
		return "", err
	}
	var objmap map[string]json.RawMessage
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	json.Unmarshal(body, &objmap)
	if string(objmap["success"]) == "false" {
		return "", errors.New("Login failed")
	}

	// Parse SID from response
	var authenticateResponse models.AuthenticateResponse
	json.Unmarshal(objmap["data"], &authenticateResponse)

	credential.SetSID(authenticateResponse.Sid)
	a.BaseApi.ApiCredential = &credential
	return authenticateResponse.Sid, nil
}