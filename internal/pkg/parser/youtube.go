package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type YoutubeParser struct{}

func (p YoutubeParser) Parse(uri string) (Content, error) {
	var content Content

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return content, InvalidURLError
	}
	if u.Host != HOST_YOUTUBE {
		return content, fmt.Errorf("Invalid youtube url %s", u.String())
	}
	return downloadSub(uri)
}

type SubTitleParser struct {
	Title       string `json:"title"`
	Duration    string `json:"duration"`
	DurationRaw string `json:"duration_raw"`
	Uploader    string `json:"uploader"`
	Thumbnail   string `json:"thumbnail"`
	Type        string `json:"type"`
	Request     struct {
		Url    string `json:"url"`
		Hash   string `json:"hash"`
		Domain string `json:"domain"`
	} `json:"request"`
	Formats []struct {
		Quality string `json:"quality"`
		Url     string `json:"url"`
		Code    string `json:"code"`
		Ext     string `json:"ext"`
	} `json:"formats"`
}

func downloadSub(uri string) (Content, error) {
	var content Content
	sub, err := fetchSubMeta(uri)
	if err != nil {
		return content, err
	}
	if len(sub.Formats) == 0 {
		return content, fmt.Errorf("no subtitle found")
	}
	subtitle, err := download(sub)
	if err != nil {
		return content, err
	}
	return Content{
		URL:     uri,
		Title:   sub.Title,
		Content: string(subtitle),
	}, nil
}

func fetchSubMeta(uri string) (*SubTitleParser, error) {
	body := `{"data":{"url":"` + uri + `"}}`
	req, err := http.NewRequest(http.MethodPost, "https://savesubs.com/action/extract", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Auth-Token", "p5Wqm6XYlaRimKelaJhoasdrlZppbGxqlpVrlWyZnmxgkJmG3mzdtK6JroKirYuEg4eKkZOxc4yp")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-Domain", "savesubs.com")
	req.Header.Add("X-Requested-With", "xmlhttprequest")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	var data struct {
		Response SubTitleParser `json:"response"`
	}
	// var data any
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	slog.Debug("get sub", "data", data)

	return &data.Response, nil
}

func download(sub *SubTitleParser) ([]byte, error) {
	params := url.Values{}
	params.Add("fileName", sub.Title)
	params.Add("ext", "txt")
	params.Add("stripAngle", "false")
	params.Add("stripParentheses", "false")
	params.Add("stripCurly", "false")
	params.Add("stripSquare", "false")
	params.Add("stripMusicCues", "false")

	u := url.URL{
		Scheme:   "https",
		Host:     "savesubs.com",
		Path:     sub.Formats[0].Url,
		RawQuery: params.Encode(),
	}

	resp, err := http.DefaultClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
