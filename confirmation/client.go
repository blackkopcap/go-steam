package confirmation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	authenticator "github.com/bbqtd/go-steam-authenticator"
	"github.com/vuquang23/go-steam/community"
)

type Client struct {
	client *http.Client

	sessionID      string
	identitySecret string

	timeOffset int64
	deviceID   string
	steamID    string
}

func NewClient(
	sessionID string,
	deviceID string,
	identitySecret string,
	steamID string,
) *Client {
	c := Client{
		client:         new(http.Client),
		sessionID:      sessionID,
		identitySecret: identitySecret,
		timeOffset:     0,
		deviceID:       deviceID,
		steamID:        steamID,
	}

	// if err := c.UpdateTimeOffset(); err != nil {
	// 	return nil, err
	// }

	return &c
}

func (c *Client) SetCookies(cookies []*http.Cookie) error {
	return community.SetCookies(c.client, cookies)
}

func (c *Client) UpdateTimeOffset() error {
	req, err := http.NewRequest(http.MethodPost, serverTimeUrl, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var body struct {
		Response struct {
			ServerTime string `json:"server_time"`
		} `json:"response"`
	}
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		return err
	}

	st, err := strconv.ParseUint(body.Response.ServerTime, 10, 64)
	if err != nil {
		return err
	}

	c.timeOffset = int64(st) - time.Now().Unix()

	return nil
}

func (c *Client) GetConfirmations() ([]*Confirmation, error) {
	// call steam server
	req := "getlist"
	key, err := c.generateConfirmationCode(loadConfirmationTag)
	if err != nil {
		return nil, err
	}
	resBytes, err := c.call(req, key, loadConfirmationTag, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		Success bool            `json:"success"`
		Conf    []*Confirmation `json:"conf"`
	}
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		return nil, err
	}

	return res.Conf, nil
}

func (c *Client) GetOfferID(conf *Confirmation) (uint64, error) {
	req := fmt.Sprintf("%s/%s", "detailspage", conf.ID)
	key, err := c.generateConfirmationCode(tradeInfoTag)
	if err != nil {
		return 0, err
	}
	resBytes, err := c.call(req, key, tradeInfoTag, nil)
	if err != nil {
		return 0, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resBytes))
	if err != nil {
		return 0, err
	}

	offer := doc.Find(".tradeoffer")
	if offer == nil {
		return 0, ErrCannotFindOffer
	}

	value, ok := offer.Attr("id")
	if !ok {
		return 0, ErrCannotFindOffer
	}
	strs := strings.Split(value, "_")
	if len(strs) < 2 {
		return 0, ErrCannotFindOffer
	}

	offerID, err := strconv.ParseUint(strs[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return offerID, nil
}

func (c *Client) AcceptConfirmation(conf *Confirmation) error {
	return c.AnswerConfirmation(conf, acceptTradeTag)
}

func (c *Client) CancelConfirmation(conf *Confirmation) error {
	return c.AnswerConfirmation(conf, cancelTag)
}

func (c *Client) AnswerConfirmation(conf *Confirmation, tag string) error {
	key, err := c.generateConfirmationCode(tag)
	if err != nil {
		return err
	}

	req := "ajaxop"
	values := jsonObj{
		"op":  tag,
		"cid": conf.ID,
		"ck":  conf.Nonce,
	}
	bytes, err := c.call(req, key, tag, values)
	if err != nil {
		return err
	}

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	return nil
}

func (c *Client) call(req string, key string, tag string, values jsonObj) ([]byte, error) {
	params := url.Values{
		"p":   {c.deviceID},
		"a":   {c.steamID},
		"k":   {key},
		"t":   {strconv.FormatUint(uint64(time.Now().Unix()+c.timeOffset), 10)},
		"m":   {"android"},
		"tag": {tag},
	}
	if len(values) != 0 {
		for k, v := range values {
			switch v := v.(type) {
			case string:
				params.Add(k, v)
			case uint64:
				params.Add(k, strconv.FormatUint(v, 10))
			default:
				return nil, fmt.Errorf("type %v is unsupported", v)
			}
		}
	}

	path := fmt.Sprintf("%s/%s?%s", "https://steamcommunity.com/mobileconf", req, params.Encode())
	request, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", defaultUserAgent)
	request.Header.Set("Accept", "*/*")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (c *Client) generateConfirmationCode(tag string) (string, error) {
	timer := func() uint64 {
		return uint64(time.Now().Unix() + c.timeOffset)
	}
	switch tag {
	case acceptTradeTag:
		return authenticator.GenerateAcceptTradeCode(c.identitySecret, timer)
	case cancelTag:
		return authenticator.GenerateCancelCode(c.identitySecret, timer)
	case loadConfirmationTag:
		return authenticator.GenerateLoadConfirmationCode(c.identitySecret, timer)
	case tradeInfoTag:
		return authenticator.GenerateTradeInfoCode(c.identitySecret, timer)
	default:
		return "", errors.New("invalid tag")
	}
}
