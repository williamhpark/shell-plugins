package ibmcloud

import (
	"context"
	"fmt"
	"os"

	"github.com/1Password/shell-plugins/sdk"
	"github.com/1Password/shell-plugins/sdk/importer"
	"github.com/1Password/shell-plugins/sdk/provision"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/1Password/shell-plugins/sdk/schema/credname"
	"github.com/1Password/shell-plugins/sdk/schema/fieldname"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/authentication/uaa"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/rest"
)

func APIKey() schema.CredentialType {
	return schema.CredentialType{
		Name:          credname.APIKey,
		DocsURL:       sdk.URL("https://cloud.ibm.com/docs/account?topic=account-userapikey"),
		ManagementURL: sdk.URL("https://cloud.ibm.com/iam/apikeys"),
		Fields: []schema.CredentialField{
			{
				Name:                fieldname.APIKey,
				MarkdownDescription: "API Key used to authenticate to IBM Cloud.",
				Secret:              true,
				Composition: &schema.ValueComposition{
					Length: 44,
					Charset: schema.Charset{
						Uppercase: true,
						Lowercase: true,
						Digits:    true,
					},
				},
			},
		},
		// DefaultProvisioner: provision.EnvVars(defaultEnvVarMapping),
		DefaultProvisioner: provision.TempFile(ibmcloudConfig, provision.Filename("config.json"), provision.AtFixedPath("~/.op/plugins/ibmcloud")),
		Importer: importer.TryAll(
			importer.TryEnvVarPair(defaultEnvVarMapping),
			TryIBMCloudConfigFile(),
		)}
}

var defaultEnvVarMapping = map[string]sdk.FieldName{
	// Set IBMCLOUD_HOME to a temp directory where we can write config.json?
	// "IBMCLOUD_API_KEY": fieldname.APIKey,
	"IBMCLOUD_HOME": "~/.op/plugins/ibmcloud",
}

func ibmcloudConfig(in sdk.ProvisionInput) ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := homeDir + "/.bluemix/config.json"
	fmt.Println(configDir)
	configFile, err := os.ReadFile(homeDir + "/.bluemix/config.json")
	if err != nil {
		return nil, err
	}
	content := configFile

	tokenRequest := uaa.APIKeyTokenRequest(fieldname.APIKey.String())
	restClient := uaa.NewClient(uaa.DefaultConfig("https://cloud.ibm.com"), rest.NewClient())
	token, err := restClient.GetToken(tokenRequest)
	if err != nil {
		return nil, err
	}
	accessToken := token.AccessToken
	refreshToken := token.RefreshToken
	expiry := token.Expiry

	fmt.Println(accessToken, " + ", refreshToken, " + ", expiry)

	// Write access token and refresh token to config file
	// Write access token, refresh token and expiry to encrypted cache?
	// Or write whole config file to encrypted cache?
	return []byte(content), nil
}

func TryIBMCloudConfigFile() sdk.Importer {
	return importer.TryFile("~/.bluemix/config.json",
		func(ctx context.Context, contents importer.FileContents,
			in sdk.ImportInput, out *sdk.ImportAttempt) {
			var config Config
			if err := contents.ToJSON(&config); err != nil {
				out.AddError(err)
				return
			}

			if config.APIKey == "" {
				return
			}

			out.AddCandidate(sdk.ImportCandidate{
				Fields: map[sdk.FieldName]string{
					fieldname.APIKey: config.APIKey,
				},
			})
		})
}

type Config struct {
	APIKey string `json:"APIKey"`
}
