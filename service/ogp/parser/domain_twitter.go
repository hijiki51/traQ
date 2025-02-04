package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dyatlov/go-opengraph/opengraph"
	jsonIter "github.com/json-iterator/go"
)

type TwitterSyndicationAPIResponse struct {
	Text string `json:"text"`
	User struct {
		Name            string `json:"name"`
		ProfileImageURL string `json:"profile_image_url_https"`
		ScreenName      string `json:"screen_name"`
	} `json:"user"`
	Photos []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"photos"`
	Video struct {
		Poster string `json:"poster"`
	} `json:"video"`
}

func FetchTwitterInfo(url *url.URL) (*opengraph.OpenGraph, *DefaultPageMeta, error) {
	splitPath := strings.Split(url.Path, "/")
	if len(splitPath) < 4 || splitPath[2] != "status" {
		return nil, nil, ErrDomainRequest
	}
	statusID := splitPath[3]
	apiResponse, err := fetchTwitterSyndicationAPI(statusID)
	if err != nil {
		return nil, nil, err
	}
	og := opengraph.OpenGraph{
		Title:       fmt.Sprintf("%s on Twitter", apiResponse.User.Name),
		Description: apiResponse.Text,
		URL:         url.String(),
	}
	result := DefaultPageMeta{}
	if len(apiResponse.Photos) > 0 {
		image := apiResponse.Photos[0]
		og.Images = []*opengraph.Image{{
			URL:    image.URL,
			Width:  uint64(image.Width),
			Height: uint64(image.Height),
		}}
	} else if apiResponse.Video.Poster != "" {
		og.Images = []*opengraph.Image{{
			URL: apiResponse.Video.Poster,
		}}
	} else if apiResponse.User.ProfileImageURL != "" {
		og.Images = []*opengraph.Image{{
			URL: apiResponse.User.ProfileImageURL,
		}}
	}
	return &og, &result, nil
}

func fetchTwitterSyndicationAPI(statusID string) (*TwitterSyndicationAPIResponse, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	requestURL := fmt.Sprintf("https://syndication.twitter.com/tweet?id=%s", statusID)
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, ErrNetwork
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return nil, ErrServer
	} else if resp.StatusCode >= 400 {
		return nil, ErrClient
	}

	data := TwitterSyndicationAPIResponse{}
	if err = jsonIter.ConfigFastest.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
