package repositories

import (
	"net/http"

	"github.com/pivotal-cf/credhub-cli/errors"

	"encoding/json"

	"github.com/pivotal-cf/credhub-cli/client"
	"github.com/pivotal-cf/credhub-cli/models"

	cm_errors "github.com/pivotal-cf/credhub-cli/errors"
)

type secretQueryRepository struct {
	httpClient client.HttpClient
}

func NewSecretQueryRepository(httpClient client.HttpClient) Repository {
	return secretQueryRepository{httpClient: httpClient}
}

func (r secretQueryRepository) SendRequest(request *http.Request, identifier string) (interface{}, error) {
	response, err := DoSendRequest(r.httpClient, request)
	if err != nil {
		return models.SecretQueryResponseBody{}, err
	}

	decoder := json.NewDecoder(response.Body)
	findResponseBody := models.SecretQueryResponseBody{}
	err = decoder.Decode(&findResponseBody)
	if err != nil {
		return models.SecretQueryResponseBody{}, cm_errors.NewResponseError()
	} else if len(findResponseBody.Credentials) < 1 {
		return models.SecretQueryResponseBody{}, errors.NewNoMatchingCredentialsFoundError()
	}
	return findResponseBody, nil
}