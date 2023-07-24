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

	var (
		getSent              = true
		getReceived          = true
		getDescriptions      = false
		activeOnly           = true
		historicalOnly       = false
		timeHistoricalCutoff = uint32(0) // https://dev.doctormckay.com/topic/4013-ieconservicegettradeoffersv1-time_historical_cutoff-on-sent-offers/
	)

	offers, err := client.GetOffers(getSent, getReceived, getDescriptions, activeOnly, historicalOnly, &timeHistoricalCutoff)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.Marshal(offers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(bytes))
}
