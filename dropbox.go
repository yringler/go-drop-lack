package lackdr

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type TemporaryLinkReponse struct {
	Link string `json:"link"`
}

// AccessToken is the token which will be used to authenticate dropbox requests.
var AccessToken = os.Getenv("dropbox_token")

// GetShareLink returns a temporary, 3 hour share link to the given URL.
func GetShareLink(dropboxPath string) (string, error) {
	response, err := makeDropRequest("https://api.dropboxapi.com/2/files/get_temporary_link", "{\"path\":\""+dropboxPath+"\"}")
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New("Failed link generation: status: " + response.Status)
	}

	bodyText, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var shareURL TemporaryLinkReponse
	err = json.Unmarshal(bodyText, &shareURL)
	if err != nil {
		return "", err
	}

	return shareURL.Link, nil
}

func makeDropRequest(url string, body string) (*http.Response, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+accessToken)
	request.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
