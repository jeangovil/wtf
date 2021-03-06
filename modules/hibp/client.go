package hibp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	apiURL            = "https://haveibeenpwned.com/api/breachedaccount/"
	apiVersion        = "application/vnd.haveibeenpwned.v2+json"
	clientTimeoutSecs = 2
	userAgent         = "WTFUtil"
)

func (widget *Widget) fullURL(account string, truncated bool) string {
	truncStr := "false"
	if truncated == true {
		truncStr = "true"
	}

	return apiURL + account + fmt.Sprintf("?truncateResponse=%s", truncStr)
}

func (widget *Widget) fetchForAccount(account string, since string) (*Status, error) {
	if account == "" {
		return nil, nil
	}

	hibpClient := http.Client{
		Timeout: time.Second * clientTimeoutSecs,
	}

	asTruncated := true
	if since != "" {
		asTruncated = false
	}

	request, err := http.NewRequest(http.MethodGet, widget.fullURL(account, asTruncated), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", apiVersion)
	request.Header.Set("User-Agent", userAgent)

	response, getErr := hibpClient.Do(request)
	if getErr != nil {
		return nil, err
	}

	body, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, err
	}

	stat, err := widget.parseResponseBody(account, body)
	if err != nil {
		return nil, err
	}

	return stat, nil
}

func (widget *Widget) parseResponseBody(account string, body []byte) (*Status, error) {
	// If the body is empty then there's no breaches
	if len(body) == 0 {
		stat := NewStatus(account, []Breach{})
		return stat, nil
	}

	// Else we have breaches for this account
	breaches := make([]Breach, 0)

	jsonErr := json.Unmarshal(body, &breaches)
	if jsonErr != nil {
		return nil, jsonErr
	}

	breaches = widget.filterBreaches(breaches)

	return NewStatus(account, breaches), nil
}

func (widget *Widget) filterBreaches(breaches []Breach) []Breach {
	// If there's no valid since value in the settings, there's no point in trying to filter
	// the breaches on that value, they'll all pass
	if !widget.settings.HasSince() {
		return breaches
	}

	sinceDate, err := widget.settings.SinceDate()
	if err != nil {
		return breaches
	}

	latestBreaches := []Breach{}

	for _, breach := range breaches {
		breachDate, err := breach.BreachDate()
		if err != nil {
			// Append the erring breach here because a failing breach date doesn't mean that
			// the breach itself isn't applicable. The date could be missing or malformed,
			// in which case we err on the side of caution and assume that the breach is valid
			latestBreaches = append(latestBreaches, breach)
			continue
		}

		if breachDate.After(sinceDate) {
			latestBreaches = append(latestBreaches, breach)
		}
	}

	return latestBreaches
}
