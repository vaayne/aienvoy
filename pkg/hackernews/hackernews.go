/*
hackernews package contains the logic to fetch the top stories from hackernews
https://github.com/HackerNews/API
*/
package hackernews

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultHost = "https://hacker-news.firebaseio.com"
	UserPath    = "/v0/user/%s.json"
	ItemPath    = "/v0/item/%d.json"
	MaxPath     = "/v0/maxitem.json"
	TopPath     = "/v0/topstories.json"
	NewPath     = "/v0/newstories.json"
	BestPath    = "/v0/beststories.json"
	JobPath     = "/v0/jobstories.json"
	AskPath     = "/v0/askstories.json"
	ShowPath    = "/v0/showstories.json"
	UpdatePath  = "/v0/updates.json"
	HNHost      = "https://news.ycombinator.com"
)

type Client struct {
	BaseUrl string
}

func New() *Client {
	return &Client{
		BaseUrl: DefaultHost,
	}
}

func (c *Client) get(path string, obj any) error {
	uri := c.BaseUrl + path
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("Error creating request for url: %s, error: %w", uri, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request for url: %s, error: %w", uri, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("Error making request for url: %s, status code: %d", uri, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response for url: %s, error: %w", uri, err)
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, obj)
	if err != nil {
		return fmt.Errorf("Error decoding response for url: %s, error: %w", uri, err)
	}
	return nil
}

// GetItem returns an Item struct with the information corresponding to the item with the provided id
func (c *Client) GetItem(id int) (Item, error) {
	var item Item
	err := c.get(fmt.Sprintf(ItemPath, id), &item)
	return item, err
}

// GetUser returns a User struct with the information of a user corresponding to the provided username
func (c *Client) GetUser(username string) (User, error) {
	var user User
	err := c.get(fmt.Sprintf(UserPath, username), &user)
	return user, err
}

// GetStory returns a Story struct with the information of a story corresponding to the provided id
func (c *Client) GetStory(id int) (Story, error) {
	item, err := c.GetItem(id)
	var story Story
	if err != nil {
		return story, err
	}
	if item.Type != "story" {
		return story, fmt.Errorf("Item with id %d is not a story", id)
	}
	return item.ToStory(), nil
}

// GetComment returns a Comment struct with the information of a comment corresponding to the provided id
func (c *Client) GetComment(id int) (Comment, error) {
	item, err := c.GetItem(id)
	var comment Comment
	if err != nil {
		return comment, err
	}
	if item.Type != "comment" {
		return comment, fmt.Errorf("Item with id %d is not a comment", id)
	}
	return item.ToComment(), nil
}

func (c *Client) GetAsk(id int) (Ask, error) {
	item, err := c.GetItem(id)
	var ask Ask
	if err != nil {
		return ask, err
	}
	if item.Type != "ask" {
		return ask, fmt.Errorf("Item with id %d is not a ask", id)
	}
	return item.ToAsk(), nil
}

// GetComment returns a Comment struct with the information of a comment corresponding to the provided id
func (c *Client) GetJob(id int) (Job, error) {
	item, err := c.GetItem(id)
	var job Job
	if err != nil {
		return job, err
	}
	if item.Type != "job" {
		return job, fmt.Errorf("Item with id %d is not a job", id)
	}
	return item.ToJob(), nil
}

// GetPoll returns a Poll struct with the information of a poll corresponding to the provided id
func (c *Client) GetPoll(id int) (Poll, error) {
	item, err := c.GetItem(id)
	var poll Poll
	if err != nil {
		return poll, err
	}
	if item.Type != "poll" {
		return poll, fmt.Errorf("Item with id %d is not a poll", id)
	}
	return item.ToPoll(), nil
}

// GetPollOpt returns a Poll struct with the information of a poll corresponding to the provided id
func (c *Client) GetPollOpt(id int) (PollOpt, error) {
	item, err := c.GetItem(id)
	var pollopt PollOpt
	if err != nil {
		return pollopt, err
	}
	if item.Type != "pollopt" {
		return pollopt, fmt.Errorf("Item with id %d is not a pollopt", id)
	}
	return item.ToPollOpt(), nil
}

func (c *Client) getStories(path string, limit, hardLimit int) ([]int, error) {
	var items []int
	if limit > hardLimit {
		return items, fmt.Errorf("limit %d greater than maximum %d items allowed", limit, hardLimit)
	}

	err := c.get(path, &items)
	if err != nil {
		return nil, err
	}
	return items[:limit], nil
}

// GetTopStories Up to 500 top and new stories are at /v0/topstories (also contains jobs) and /v0/newstories. Best stories are at /v0/beststories.
func (c *Client) GetTopStories(limit int) ([]int, error) {
	return c.getStories(TopPath, limit, 500)
}

// GetNewStories Up to 500 top and new stories are at /v0/topstories (also contains jobs) and /v0/newstories. Best stories are at /v0/beststories.
func (c *Client) GetNewStories(limit int) ([]int, error) {
	return c.getStories(NewPath, limit, 500)
}

// GetBestStories Up to 500 top and new stories are at /v0/topstories (also contains jobs) and /v0/newstories. Best stories are at /v0/beststories.
func (c *Client) GetBestStories(limit int) ([]int, error) {
	return c.getStories(BestPath, limit, 500)
}

// GetAskStories Up to 200 of the latest Ask HN, Show HN, and Job stories are at /v0/askstories, /v0/showstories, and /v0/jobstories.
func (c *Client) GetAskStories(limit int) ([]int, error) {
	return c.getStories(AskPath, limit, 200)
}

// GetShowStories Up to 200 of the latest Ask HN, Show HN, and Job stories are at /v0/askstories, /v0/showstories, and /v0/jobstories.
func (c *Client) GetShowStories(limit int) ([]int, error) {
	return c.getStories(ShowPath, limit, 200)
}

// GetJobStories Up to 200 of the latest Ask HN, Show HN, and Job stories are at /v0/askstories, /v0/showstories, and /v0/jobstories.
func (c *Client) GetJobStories(limit int) ([]int, error) {
	return c.getStories(JobPath, limit, 200)
}

// GetRecentChanges The item and profile changes are at /v0/updates.
func (c *Client) GetRecentChanges() (Changes, error) {
	var changes Changes
	err := c.get(UpdatePath, &changes)
	return changes, err
}

// GetMaxId The current largest item id is at /v0/maxitem. You can walk backward from here to discover all items.
func (c *Client) GetMaxId() (int, error) {
	var id int
	err := c.get(MaxPath, &id)
	return id, err
}
