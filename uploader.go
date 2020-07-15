package p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

const googlePhotoUploadAPIPath = "https://photoslibrary.googleapis.com/v1/uploads"
const googlePhotoMediaItemsCreateAPIPath = "https://photoslibrary.googleapis.com/v1/mediaItems:batchCreate"
const gooogleTokenAPIPath = "https://www.googleapis.com/oauth2/v4/token"

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

var tokenPool struct {
	expiredAt time.Time
	token     string
}

// GooglePhotoUploader uploads image binaries to an album
type GooglePhotoUploader struct {
	albumID string
	token   string
}

// NewGooglePhotoUploader is a constructor
func NewGooglePhotoUploader(config *config) (*GooglePhotoUploader, error) {
	token, err := getToken(config)
	if err != nil {
		return nil, err
	}
	return &GooglePhotoUploader{
		albumID: config.AlbumID,
		token:   token,
	}, nil
}

func getToken(cfg *config) (string, error) {
	now := time.Now()
	if now.Before(tokenPool.expiredAt) {
		return tokenPool.token, nil
	}

	form := url.Values{}
	form.Add("refresh_token", cfg.Oauth2RefreshToken)
	form.Add("client_id", cfg.Oauth2ClientID)
	form.Add("client_secret", cfg.Oauth2ClientSecret)
	form.Add("grant_type", "refresh_token")
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest(http.MethodPost, gooogleTokenAPIPath, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", xerrors.Errorf("status:%s body=%s", resp.Status, bodyBytes)
	}

	var jsonResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return "", err
	}

	tokenPool.token = jsonResp.AccessToken
	tokenPool.expiredAt = time.Unix(now.Unix()+jsonResp.ExpiresIn, 0)
	outputLog("token generated. token=%s scope = %s, expiredAt = %s", jsonResp.AccessToken, jsonResp.Scope, tokenPool.expiredAt.Format(time.RFC3339))

	return tokenPool.token, nil

}

func (u *GooglePhotoUploader) uploadContent(ctx context.Context, content io.Reader) (uploadToken string, err error) {
	var contentBytes []byte
	if _, err := content.Read(contentBytes); err != nil {
		return "", xerrors.Errorf("Failed to read bytes: %w", err)
	}
	mimeType := http.DetectContentType(contentBytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googlePhotoUploadAPIPath, content)
	if err != nil {
		return "", xerrors.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.token))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-Content-Type", mimeType)
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", xerrors.Errorf("Failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", xerrors.Errorf("Failed to read response body: %w", err)
		}
		return "", xerrors.Errorf("Failed to upload image: detail=%s", bodyBytes)
	}

	uploadTokenBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", xerrors.Errorf("Failed to read body: %w", err)
	}
	uploadToken = string(uploadTokenBytes)
	return
}

func (u *GooglePhotoUploader) createMediaItem(ctx context.Context, uploadToken, displayName string) error {

	data := map[string]interface{}{
		"albumId": u.albumID,
		"newMediaItems": []interface{}{
			map[string]interface{}{
				"description": fmt.Sprintf("%sさんがアップロード", displayName),
				"simpleMediaItem": map[string]string{
					"uploadToken": uploadToken,
				},
			},
		},
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return xerrors.Errorf("Failed to parse: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googlePhotoMediaItemsCreateAPIPath, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return xerrors.Errorf("Failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return xerrors.Errorf("Failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return xerrors.Errorf("Failed to read response body: %w", err)
		}
		return xerrors.Errorf("Failed to add image to the album: detail=%s", bodyBytes)
	}

	return nil
}

func (u *GooglePhotoUploader) upload(ctx context.Context, displayName string, content io.Reader) error {

	uploadToken, err := u.uploadContent(ctx, content)
	if err != nil {
		return err
	}

	if err := u.createMediaItem(ctx, uploadToken, displayName); err != nil {
		return err
	}

	return nil
}
