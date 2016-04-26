package chromiumpreload

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
)

const (
	ForceHTTPS = "force-https"
)

// PreloadList contains a parsed form of the Chromium Preload list.
//
// The full list contains information about more than just HSTS, but only
// HSTS-related contents are currently exposed in this struct.
type PreloadList struct {
	Entries []PreloadEntry `json:"entries"`
}

type Domain string

// A PreloadEntry contains the data from an entry in the Chromium
// Preload list.
//
// - Name: The domain name.
//
// - Mode: The only valid non-empty value is ForceHTTPS
//
// - IncludeSubDomains: If Mode == ForceHTTPS, forces HSTS to apply to
//   all subdomains.
type PreloadEntry struct {
	Name              Domain `json:"name"`
	Mode              string `json:"mode"`
	IncludeSubDomains bool   `json:"include_subdomains"`
}

const (
	latestChromiumListURL = "https://chromium.googlesource.com/chromium/src/+/master/net/http/transport_security_state_static.json?format=TEXT"
)

// GetLatest retrieves the latest PreloadList from the Chromium source at
// https://chromium.googlesource.com/chromium/src/+/master/net/http/transport_security_state_static.json
//
// Note that this list may be up to 12 weeks fresher than the list used
// by the current stable version of Chrome. See
// https://www.chromium.org/developers/calendar for a calendar of releases.
func GetLatest() (PreloadList, error) {
	var preloadList PreloadList

	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(latestChromiumListURL)
	if err != nil {
		return preloadList, err
	}

	if resp.StatusCode != 200 {
		return preloadList, fmt.Errorf("Status code %d", resp.StatusCode)
	}

	body := base64.NewDecoder(base64.StdEncoding, resp.Body)
	jsonBytes, err := removeComments(body)
	if err != nil {
		return preloadList, errors.New("Could not decode body.")
	}

	if err := json.Unmarshal(jsonBytes, &preloadList); err != nil {
		return preloadList, err
	}

	return preloadList, nil
}

// PreloadEntriesToMap creates an indexed map (Domain -> PreloadEntry) of
// the entries from the given PreloadList.
func PreloadEntriesToMap(preloadList PreloadList) map[Domain]PreloadEntry {
	m := make(map[Domain]PreloadEntry)
	for _, entry := range preloadList.Entries {
		m[entry.Name] = entry
	}
	return m
}

// removeComments reads the contents of |r| and removes any lines beginning
// with optional whitespace followed by "//"
func removeComments(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if isCommentLine(line) {
			fmt.Fprintln(&buf, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func isCommentLine(line string) bool {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	return !strings.HasPrefix(trimmed, "//")
}
