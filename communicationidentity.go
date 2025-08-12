// Unofficial client library for REST API calls to 'Azure Communication Identity' routes
// on a given 'Azure Communication Services' endpoint (WIP).
//
// The main entry point is [communicationidentity.New] to create a new [communicationidentity.CommunicationIdentityClient],
// check the examples section for more guidance.
package communicationidentity

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// REST client to perform calls to 'Azure Communication Identity' endpoints
// on a given 'Azure Communication Services' instance
type CommunicationIdentityClient struct {
	acsEndpoint      *url.URL
	decodedAcsSecret []byte
	azClientId       string
}

type azAPIVersion string

const (
	tokenForTeamsUserEndpoint               = "/teamsUser/:exchangeAccessToken"
	createUserAndTokenEndpoint              = "/identities"
	apiVersion                 azAPIVersion = "2025-06-30"
	msAuthHeader                            = "Authorization"
	msDateHeader                            = "x-ms-date"
	msContentHashHeader                     = "x-ms-content-sha256"
)

// constructor for the REST Client
func New(
	acsEndpoint *url.URL,
	acsAccessKey string,
	azClientId string,
) (CommunicationIdentityClient, error) {
	decodedAcsSecret, err := base64.StdEncoding.DecodeString(acsAccessKey)
	if err != nil {
		return CommunicationIdentityClient{}, fmt.Errorf(
			"ACS access key is not valid base64: %w",
			err,
		)
	}
	return CommunicationIdentityClient{acsEndpoint, decodedAcsSecret, azClientId}, nil
}

func (client CommunicationIdentityClient) buildEndpointURL(
	endpoint string,
	apiVersion azAPIVersion,
) *url.URL {
	endpointURL := client.acsEndpoint.JoinPath(endpoint)
	query := endpointURL.Query()
	query.Set("api-version", string(apiVersion))
	endpointURL.RawQuery = query.Encode()

	return endpointURL
}

// see: https://learn.microsoft.com/en-us/azure/communication-services/tutorials/hmac-header-tutorial?pivots=programming-language-csharp
func (client CommunicationIdentityClient) buildSignedRequest(
	url *url.URL,
	body []byte,
) (*http.Request, error) {
	if url == nil {
		return nil, fmt.Errorf("url for signed request can not be nil")
	}
	computeHash := func(content []byte) string {
		hash := sha256.Sum256(content)
		return base64.StdEncoding.EncodeToString(hash[:])
	}
	computeSignature := func(toSign string) (string, error) {
		if !utf8.ValidString(toSign) {
			return "", fmt.Errorf("string to sign is not valid utf-8")
		}

		mac := hmac.New(sha256.New, client.decodedAcsSecret)
		_, err := mac.Write([]byte(toSign))
		if err != nil {
			return "", fmt.Errorf("failed to write to MAC: %w", err)
		}
		macSum := mac.Sum(nil)

		return base64.StdEncoding.EncodeToString(macSum), nil
	}

	// DO NOT USE 'time.RFC1123' : https://github.com/golang/go/issues/13781
	date := time.Now().UTC().Format(http.TimeFormat)
	contentHash := computeHash(body)
	pathAndQuery := fmt.Sprintf("%s?%s", url.EscapedPath(), url.RawQuery)

	stringToSign := fmt.Sprintf("POST\n%s\n%s;%s;%s", pathAndQuery, date, url.Host, contentHash)
	signature, err := computeSignature(stringToSign)
	if err != nil {
		return nil, fmt.Errorf("failed to build request signature: %w", err)
	}

	authorization :=
		fmt.Sprintf(
			"HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature=%s",
			signature,
		)

	request, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(msDateHeader, date)
	request.Header.Add(msContentHashHeader, contentHash)
	request.Header.Add(msAuthHeader, authorization)

	return request, nil
}

type teamsUserExchangeTokenRequest struct {
	AppId  string `json:"appId"`
	Token  string `json:"token"`
	UserId string `json:"userId"`
}

type CommunicationIdentityAccessToken struct {
	Token     string    `json:"token"`
	ExpiresOn time.Time `json:"expiresOn"`
}

type CommunicationIdentityAccessTokenResult struct {
	AccessToken CommunicationIdentityAccessToken `json:"accessToken"`
	Identity    struct {
		ID string `json:"id"`
	} `json:"identity"`
}

// Machine-readable errors returned from Azure Communication Services endpoints.
// `Code` can be used to handle errors in a stable way, though microsoft may add
// new codes in the future
//
// NOTE: no Unwrap implementation to Innererror on purpose, this may change
type CommunicationError struct {
	Code       string               `json:"code"`
	Details    []CommunicationError `json:"details"`
	Innererror *CommunicationError  `json:"innererror"`
	Message    string               `json:"message"`
	// TODO: find usecases of this and improve formatted output accordingly
	Target string `json:"target"`
}

func (err *CommunicationError) Error() string {
	var out strings.Builder

	if err.Target != "" {
		out.WriteString(fmt.Sprintf("[target:%s]", err.Target))
	}
	out.WriteString(
		fmt.Sprintf("%s - %s\n", err.Code, err.Message))
	out.WriteString(fmt.Sprintf("details: %+v\n", err.Details))
	if err.Innererror != nil {
		out.WriteString(fmt.Sprintf("inner error: %v\n", err.Innererror))
	}
	return out.String()
}

type communicationErrorResponse struct {
	Error CommunicationError `json:"error"`
}

// Azure Documentation: https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/exchange-teams-user-access-token?view=rest-communication-identity-2025-06-30&tabs=HTTP
func (client CommunicationIdentityClient) TokenForTeamsUser(
	ctx context.Context,
	userOid string,
	teamsScopeMSALToken string,
) (CommunicationIdentityAccessToken, error) {
	fullResourceURL := client.buildEndpointURL(tokenForTeamsUserEndpoint, apiVersion)
	requestBody, err := json.Marshal(teamsUserExchangeTokenRequest{
		AppId:  client.azClientId,
		Token:  teamsScopeMSALToken,
		UserId: userOid,
	})
	if err != nil {
		return CommunicationIdentityAccessToken{}, fmt.Errorf(
			"failed to build request body: %w",
			err,
		)
	}
	request, err := client.buildSignedRequest(fullResourceURL, requestBody)
	if err != nil {
		return CommunicationIdentityAccessToken{}, fmt.Errorf(
			"failed to create signed request: %w",
			err,
		)
	}
	request = request.WithContext(ctx)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return CommunicationIdentityAccessToken{}, fmt.Errorf(
			"failed to send request to ACS: %w",
			err,
		)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			// TODO: do something nicer here
			fmt.Fprintf(
				os.Stderr,
				"'Communication Identity' failed to close response body: %v",
				err,
			)
		}
	}()

	if response.StatusCode == http.StatusOK {
		var tokenResponse CommunicationIdentityAccessToken
		if err := json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
			return CommunicationIdentityAccessToken{}, fmt.Errorf(
				"failed to parse response body for status OK",
			)
		}
		return tokenResponse, nil

	} else {
		var errorResponse communicationErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return CommunicationIdentityAccessToken{}, fmt.Errorf("ACS responded with non-OK status(%v) and response body was not parseable", response.Status)
		}

		return CommunicationIdentityAccessToken{}, fmt.Errorf("ACS responded with non-OK status(%v), error: %w", response.Status, &errorResponse.Error)
	}
}

type createAndReturnTokenRequest struct {
	Scope  []string `json:"createTokenWithScopes"`
	Expire *int32   `json:"expiresInMinutes,omitempty"`
}

// CreateCommunicationIdentity Azure Documentation https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/create?view=rest-communication-identity-2025-06-30&tabs=HTTP
func (client CommunicationIdentityClient) CreateCommunicationIdentity(ctx context.Context, scope []string, expireInMinutes *int32) (CommunicationIdentityAccessTokenResult, error) {
	fullResourceURL := client.buildEndpointURL(createUserAndTokenEndpoint, apiVersion)

	requestBody, err := json.Marshal(createAndReturnTokenRequest{
		Scope:  scope,
		Expire: expireInMinutes,
	})
	if err != nil {
		return CommunicationIdentityAccessTokenResult{}, fmt.Errorf(
			"failed to build requeset body: %w",
			err,
		)
	}

	request, err := client.buildSignedRequest(fullResourceURL, requestBody)

	if err != nil {
		return CommunicationIdentityAccessTokenResult{}, fmt.Errorf(
			"failed to create signed request: %w",
			err,
		)
	}
	request = request.WithContext(ctx)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return CommunicationIdentityAccessTokenResult{}, fmt.Errorf(
			"failed to send reqeust to ACS: %w",
			err,
		)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			// TODO: do something nicer here
			fmt.Fprintf(
				os.Stderr,
				"'Communication Identity' failed to close response body: %v",
				err,
			)
		}
	}()
	if response.StatusCode == http.StatusCreated {
		var tokenResponse CommunicationIdentityAccessTokenResult
		if err := json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
			return CommunicationIdentityAccessTokenResult{}, fmt.Errorf(
				"failed to parse response body for status OK",
			)
		}
		return tokenResponse, nil

	} else {
		var errorResponse communicationErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return CommunicationIdentityAccessTokenResult{}, fmt.Errorf("ACS responded with non-OK status(%v) and response body was not parseable", response.Status)
		}

		return CommunicationIdentityAccessTokenResult{}, fmt.Errorf("ACS responded with non-OK status(%v), error: %w", response.Status, &errorResponse.Error)
	}
}
