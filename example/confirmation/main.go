package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/vuquang23/go-steam/community"
	"github.com/vuquang23/go-steam/confirmation"
	"github.com/vuquang23/go-steam/totp"
)

var (
	accountName    string
	password       string
	identitySecret string
	sharedSecret   string
)

const baseUrl = "https://steamcommunity.com"

func init() {
	accountName = os.Getenv("ACCOUNT_NAME")
	password = os.Getenv("PASSWORD")
	identitySecret = os.Getenv("IDENTITY_SECRET")
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

	c := confirmation.NewClient(
		communityClient.GetSessionID(),
		communityClient.GetDeviceID(),
		identitySecret,
		communityClient.GetSteamID(),
	)

	communityUrl, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = c.SetCookies(communityClient.GetCookies(communityUrl))
	if err != nil {
		log.Fatal(err)
	}

	confs, err := c.GetConfirmations()
	if err != nil {
		log.Fatal(err)
	}
	if len(confs) == 0 {
		return
	}

	m, _ := json.Marshal(confs)
	fmt.Println(string(m))

	offerID, err := c.GetOfferID(confs[0])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(offerID)

	// err = c.AcceptConfirmation(confs[0])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err = c.CancelConfirmation(confs[0])
	if err != nil {
		log.Fatal(err)
	}
}
