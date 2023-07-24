package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/vuquang23/go-steam/community"
	"github.com/vuquang23/go-steam/totp"
	"github.com/vuquang23/go-steam/tradeoffer"
)

var (
	apiKey       string
	accountName  string
	password     string
	sharedSecret string
)

const baseUrl = "https://steamcommunity.com"

func init() {
	apiKey = os.Getenv("API_KEY")
	accountName = os.Getenv("ACCOUNT_NAME")
	password = os.Getenv("PASSWORD")
	sharedSecret = os.Getenv("SHARED_SECRET")
}

func main() {
	communityClient, err := community.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	code, err := totp.GenerateTotpCode(sharedSecret, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	err = communityClient.Login(community.LoginDetails{
		AccountName:   accountName,
		Password:      password,
		TwoFactorCode: code,
	})
	if err != nil {
		log.Fatal(err)
	}

	client := tradeoffer.NewClient(
		tradeoffer.APIKey(apiKey),
		communityClient.GetSessionID(),
	)
	communityUrl, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SetCookies(communityClient.GetCookies(communityUrl))
	if err != nil {
		log.Fatal(err)
	}

	data, err := client.GetOwnInventory(2, 730, true)
	if err != nil {
		log.Fatal(err)
	}

	m, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(m))
}
