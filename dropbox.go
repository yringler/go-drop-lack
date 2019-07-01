package lackdr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type TemporaryLinkReponse struct {
	Link string `json:"link"`
}

// AccessToken is the token which will be used to authenticate dropbox requests.
var AccessToken string

// GetShareLink returns a temporary, 3 hour share link to the given URL.
func GetShareLink(dropboxPath string) (string, error) {
	response, err := MakeDropRequest("https://api.dropboxapi.com/2/files/get_temporary_link", "{\"path\":\""+dropboxPath+"\"}")
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

// MakeDropRequest makes a basic authenticated dropbox request.
func MakeDropRequest(url string, body string) (*http.Response, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+AccessToken)
	request.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func createUploadRequest(fileName, dropboxPath string) (*http.Request, error) {
	const dropboxAPI = "https://content.dropboxapi.com/2/files/upload"

	// Reopen the file before each request.
	// When the request is made, the file is closed.
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("POST", dropboxAPI, file)

	request.Header.Add("Content-Type", "application/octet-stream")
	request.Header.Add("Authorization", "Bearer "+AccessToken)
	request.Header.Add("Dropbox-API-Arg", "{\"mode\":\"overwrite\",\"path\":\""+dropboxPath+file.Name()+"\"}")

	return request, nil
}

// UploadFile uploads file to dropbox, return the dropbox path.
func UploadFile(file, dropboxPath string) (string, error) {

	// Make request until not rate-limited
	for i := 0; i < 3; i++ {
		request, err := createUploadRequest(file, dropboxPath)
		if err != nil {
			return "", err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		retry := response.Header.Get("Retry-After")

		if retry != "" {
			retrySeconds, _ := strconv.ParseInt(retry, 10, 32)
			fmt.Println("retrying: ", retrySeconds, file)
			time.Sleep(time.Duration(retrySeconds) * time.Second)
		} else {
			body, _ := ioutil.ReadAll(response.Body)

			bodyString := string(body)
			if strings.Contains(bodyString, "error") {
				fmt.Println("upload error body:\n", string(body))
			}
			break
		}
	}

	return dropboxPath + file, nil
}
