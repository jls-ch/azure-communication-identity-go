package communicationidentity_test

import (
	"context"
	"fmt"
	"net/url"

	ci "github.com/jls-ch/azure-communication-identity-go"
)

func Example() {
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
	if err != nil {
		panic(err)
	}
	fmt.Printf("token for teams user: %#v\n", token)

	// ...
}
