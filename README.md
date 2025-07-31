# azure-communication-identity-go
Unofficial client library for REST APIs to Azure Communication Identity Services for Golang without dependencies (WIP).

This library was built with Golang 1.24, earlier versions might be able to compile the code but have not been tested.

[GoDoc](https://pkg.go.dev/github.com/jls-ch/azure-communication-identity-go)

# Example Usage

see also: [example_test.go](./example_test.go)

```go
package main

import (
	"context"
	"net/url"

	ci "github.com/jls-ch/azure-communication-identity-go"
)

func main() {
	acsURL, err := url.Parse("YOUR-ACS-ENDPOINT")
	if err != nil {
		panic(err)
	}

	client, err := ci.New(
			        acsURL,
			        "YOUR-ACS-SECRET-ACCESS-KEY",
			        "ID-OF-APP-REGISTRATION-WITH-TEAMS-PERMISSIONS")
	if err != nil {
		panic(err)
	}
	token, err := client.TokenForTeamsUser(context.TODO(), "USER-OID", "ENTRA-TOKEN-WITH-TEAMS-SCOPE")
    // ...
}

```


# Features & Roadmap

Implemented:
- HMAC request and header signing
- [Azure Communication Services errors](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/create?view=rest-communication-identity-2025-06-30&tabs=HTTP#communicationerror) 
exposed through `CommunicationError`
- API version "2025-06-30" routes:
    - [Exchange Teams User Access Token](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/exchange-teams-user-access-token?view=rest-communication-identity-2025-06-30&tabs=HTTP)

Planned:
- API version "2025-06-30" routes:
    - [Create](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/create?view=rest-communication-identity-2025-06-30&tabs=HTTP)
    - [Delete](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/delete?view=rest-communication-identity-2025-06-30&tabs=HTTP)
    - [Issue Access Token](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/issue-access-token?view=rest-communication-identity-2025-06-30&tabs=HTTP)
    - [Revoke Access Token](https://learn.microsoft.com/en-us/rest/api/communication/identity/communication-identity/revoke-access-tokens?view=rest-communication-identity-2025-06-30&tabs=HTTP)

Not Planned:
- Support for older Go versions
- Support for older Azure Communication Identity API Versions

