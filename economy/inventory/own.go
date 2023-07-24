package inventory

import (
	"fmt"
	"net/http"
	"strconv"
)

func GetPartialOwnInventory(client *http.Client, contextId uint64, appId uint32, start *uint, tradableOnly bool) (*PartialInventory, error) {
	url := fmt.Sprintf("http://steamcommunity.com/my/inventory/json/%d/%d", appId, contextId)
	if tradableOnly {
		url += "?trading=1"
	}
	if start != nil {
		url += "&start=" + strconv.FormatUint(uint64(*start), 10)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DoInventoryRequest(client, req)
}

func GetOwnInventory(client *http.Client, contextId uint64, appId uint32, tradableOnly bool) (*Inventory, error) {
	return GetFullInventory(func() (*PartialInventory, error) {
		return GetPartialOwnInventory(client, contextId, appId, nil, tradableOnly)
	}, func(start uint) (*PartialInventory, error) {
		return GetPartialOwnInventory(client, contextId, appId, &start, tradableOnly)
	})
}
