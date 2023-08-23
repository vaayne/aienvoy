package parser

// Content is the content of a url
// https://github.com/postlight/parser
/*
{
  "title": "Thunder (mascot)",
  "content": "... <p><b>Thunder</b> is the <a href=\"https://en.wikipedia.org/wiki/Stage_name\">stage name</a> for the...",
  "author": "Wikipedia Contributors",
  "date_published": "2016-09-16T20:56:00.000Z",
  "lead_image_url": null,
  "dek": null,
  "next_page_url": null,
  "url": "https://en.wikipedia.org/wiki/Thunder_(mascot)",
  "domain": "en.wikipedia.org",
  "excerpt": "Thunder Thunder is the stage name for the horse who is the official live animal mascot for the Denver Broncos",
  "word_count": 4677,
  "direction": "ltr",
  "total_pages": 1,
  "rendered_pages": 1
}
*/
type Content struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	Author        string `json:"author"`
	DatePublished string `json:"date_published"`
	LeadImageURL  string `json:"lead_image_url"`
	Dek           string `json:"dek"`
	NextPageURL   string `json:"next_page_url"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Excerpt       string `json:"excerpt"`
	WordCount     int    `json:"word_count"`
	Direction     string `json:"direction"`
	TotalPages    int    `json:"total_pages"`
	RenderedPages int    `json:"rendered_pages"`
}

type Parser interface {
	Parse(url string) (Content, error)
}
