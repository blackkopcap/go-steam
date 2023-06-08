package community

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	client *http.Client

	mu      sync.Mutex
	session loginSession
}

func NewClient() (*Client, error) {
	cookies := []*http.Cookie{
		{Name: "Steam_Language", Value: "english"},
		{Name: "timezoneOffset", Value: "0,0"},
	}
	httpClient := new(http.Client)
	SetCookies(httpClient, cookies)
	return &Client{client: httpClient}, nil
}

func (c *Client) Login(details LoginDetails) error {
	if details.AccountName == "" || details.Password == "" {
		return errors.New("missing account name or password")
	}

	getRsaKeyRes, err := c.GetRSAKey(details.AccountName)
	if err != nil {
		return err
	}
	encryptedPassword, err := EncryptPassword(getRsaKeyRes.PublickeyMod, getRsaKeyRes.PublickeyExp, details.Password)
	if err != nil {
		return err
	}

	values := url.Values{
		"captcha_text":      {""},
		"captchagid":        {"-1"},
		"emailauth":         {""},
		"emailsteamid":      {""},
		"password":          {encryptedPassword},
		"remember_login":    {"true"},
		"rsatimestamp":      {getRsaKeyRes.Timestamp},
		"twofactorcode":     {details.TwoFactorCode},
		"username":          {details.AccountName},
		"loginfriendlyname": {""},
		"donotcache":        {strconv.FormatInt(time.Now().Unix()*1000, 10)},
	}.Encode()
	request, err := http.NewRequest(http.MethodPost, doLoginUrl, strings.NewReader(values))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Add("Content-Length", strconv.Itoa(len(values)))
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	request.Header.Add("Origin", baseUrl)
	request.Header.Add("Referer", loginUrl)
	request.Header.Add("User-Agent", defaultUserAgent)
	request.Header.Add("Accept", "*/*")

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	resBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var session loginSession
	err = json.Unmarshal(resBytes, &session)
	if err != nil {
		return err
	}
	if !session.Success {
		if session.RequiresTwoFactor {
			return errors.New("requires two factor")
		}
		return errors.New(session.Message)
	}

	// generate session id.
	sessionID, err := GenerateSessionID()
	if err != nil {
		return err
	}
	session.OAuth.ID = sessionID

	// gen device ID
	session.OAuth.DeviceID = GenerateDeviceID(details.AccountName, details.Password)

	// get cookies: `steamLogin`, `steamLoginSecure`
	communityUrl, err := url.Parse(baseUrl)
	if err != nil {
		return err
	}
	for _, cookie := range c.client.Jar.Cookies(communityUrl) {
		switch cookie.Name {
		case cookieSteamLogin:
			session.OAuth.SteamLogin = cookie.Value
		case cookieSteamLoginSecure:
			session.OAuth.SteamLoginSecure = cookie.Value
		}
	}

	// set login session
	c.setSession(session)

	return nil
}

func (c *Client) setSession(session loginSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = session
}

func (c *Client) GetRSAKey(accountName string) (*getRSAKeyRes, error) {
	path := fmt.Sprintf("%s?username=%s", rsaUrl, accountName)
	request, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Referer", loginUrl)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var ret getRSAKeyRes
	err = json.Unmarshal(respBytes, &ret)
	if err != nil {
		return nil, err
	}
	if !ret.Success {
		return nil, errors.New("failed to get rsa key")
	}

	return &ret, nil
}

func (c *Client) GetSteamID() string {
	return c.session.OAuth.SteamID
}

func (c *Client) GetDeviceID() string {
	return c.session.OAuth.DeviceID
}

func (c *Client) GetCookies(u *url.URL) []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(c.client.Jar.Cookies(u)))
	for _, cookie := range c.client.Jar.Cookies(u) {
		clone := *cookie
		cookies = append(cookies, &clone)
	}
	return cookies
}

func (c *Client) GetSessionID() string {
	return c.session.OAuth.ID
}
