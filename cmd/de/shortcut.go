package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/zalando/go-keyring"
)

type Member struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	MentionName string `json:"mention_name"`
}

type storySearchResponse struct {
	Stories Stories `json:"stories"`
}

type Stories struct {
	Next  string  `json:"next"`
	Data  []Story `json:"data"`
	Total int     `json:"total"`
}

type Story struct {
	StartedAt string `json:"started_at"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	MovedAt   string `json:"moved_at"`
}

var (
	secretsService          = "declitool"
	secretsShortcutUsername = "shortcut"
)

func getShortcutApiKey() (string, bool, error) {
	apiKey, err := keyring.Get(
		secretsService,
		secretsShortcutUsername,
	)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", true, nil
		} else {
			return "", true, err
		}

	} else {
		// fmt.Println("API Key retrieved")
		return apiKey, false, nil
	}
}

func promptFor(message string) (string, error) {
	fmt.Println(message)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	value := strings.TrimSuffix(input, "\n")

	return value, nil
}

func setShortcutApiKey() (string, error) {
	apiKey, err := promptFor("Enter your shortcut API key:")
	if err != nil {
		return "", err
	}
	err = keyring.Set(
		secretsService,
		secretsShortcutUsername,
		apiKey,
	)
	if err != nil {
		return "", err
	} else {
		fmt.Println("API Key set")
		return apiKey, nil
	}
}

func authenticate() (string, error) {
	apiKey, notFound, err := getShortcutApiKey()
	if err != nil {
		return "", err
	}
	if notFound {
		fmt.Println("API Key not found")
		apiKey, err = setShortcutApiKey()
		if err != nil {
			return "", err
		}
	}
	return apiKey, nil
}

func deleteShortcutApiKey() error {
	err := keyring.Delete(
		secretsService,
		secretsShortcutUsername,
	)
	if err != nil {
		return err
	} else {
		fmt.Println("Key deleted")
		return nil
	}
}

func getMentionName(apiKey string) (string, error) {
	url := "https://api.app.shortcut.com/api/v3/member"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Shortcut-Token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var member Member
	err = json.Unmarshal(body, &member)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling member response: %s, with body: %s", err, string(body))
	}

	// fmt.Printf("Mention name for %s: %s\n", member.Name, member.MentionName)

	return member.MentionName, nil
}

func getStories(apiKey string, name string) ([]Story, error) {
	url := "https://api.app.shortcut.com/api/v3/search"
	bodyParams := strings.NewReader(`{ "detail": "slim", "page_size": 25, "query": "owner:` + name + `" }`)
	req, err := http.NewRequest("GET", url, bodyParams)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Shortcut-Token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var storyResponse storySearchResponse
	err = json.Unmarshal(body, &storyResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling story response: %s, with body: %s", err, string(body))
	}

	return storyResponse.Stories.Data, nil
}

func printStories(stories []Story, config shortcutStoryConfig) error {
	var storyHeadlines []string

	sort.Slice(stories, func(i, j int) bool {
		return stories[i].MovedAt < stories[j].MovedAt
	})

	for _, story := range stories {
		kebabCaseTitle := strings.ToLower(strings.ReplaceAll(story.Name, " ", "-"))
		gitBranch := fmt.Sprintf("sc-%d-%s", story.ID, kebabCaseTitle)
		if config.withTimes {
			storyHeadline := fmt.Sprintf("%s: %s\n", story.MovedAt, gitBranch)
			storyHeadlines = append(storyHeadlines, storyHeadline)
		} else {
			storyHeadline := fmt.Sprintf("%s\n", gitBranch)
			storyHeadlines = append(storyHeadlines, storyHeadline)
		}
	}

	var limit int
	if config.numberOfStories >= 0 && config.numberOfStories <= len(storyHeadlines) {
		limit = len(storyHeadlines) - config.numberOfStories
	} else {
		limit = 0
	}
	storyHeadlines = storyHeadlines[limit:]
	output := strings.TrimRight(strings.Join(storyHeadlines, ""), "\n")
	fmt.Println(output)

	return nil
}

func myStories(config shortcutStoryConfig) error {
	pass, err := authenticate()
	if err != nil {
		return err
	}

	name, err := getMentionName(pass)
	if err != nil {
		return err
	}

	stories, err := getStories(pass, name)
	if err != nil {
		return err
	}

	err = printStories(stories, config)
	if err != nil {
		return err
	}

	return nil
}
