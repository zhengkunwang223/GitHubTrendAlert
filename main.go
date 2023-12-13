package main

import (
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-trending"
	"github.com/eatmoreapple/openwechat"
	"github.com/robfig/cron"
	"github.com/skip2/go-qrcode"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	Languages  []string
	Repos      []string
	FriendName string
	CronSpec   string
	bot        *openwechat.Bot
)

type Config struct {
	Languages  []string `yaml:"languages"`
	Repos      []string `yaml:"repos"`
	FriendName string   `yaml:"friendName"`
	CronSpec   string   `yaml:"cronSpec"`
}

type Repo struct {
	RawName string `json:"rawName"`
	Stars   struct {
		Count int `json:"count"`
	} `json:"stars"`
}

func loadConfig() (*Config, error) {
	var config Config

	configPath := "app.yaml"

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func ConsoleQrCode(uuid string) {
	qrCode, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Medium)
	fmt.Println(qrCode.ToString(true))
	qrcode.WriteFile(qrCode.Content, qrcode.Medium, 256, "qr.png")
}

func repoExist(repoName string) bool {
	for _, name := range Repos {
		if repoName == name {
			return true
		}
	}
	return false
}

func getGitHubTrendingByLanguage(language string) {
	trend := trending.NewTrending()
	projects, err := trend.GetProjects(trending.TimeToday, language)
	if err != nil {
		fmt.Println("get github trending error", err)
		return
	}
	for index, project := range projects {
		i := index + 1
		if repoExist(project.Name) {
			_ = sendMsg(fmt.Sprintf("[%s] Github Trending %s 榜单上榜! 名次 %d 当前 star %d", project.Name, language, i, project.Stars))
		}
	}
}

func getTotalGitHubTrending() {
	response, err := http.Get("https://devo-platforms.burakkarakan.com/github.json")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	var repos []Repo
	err = json.Unmarshal(body, &repos)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	for index, repo := range repos {
		repoName := strings.ReplaceAll(repo.RawName, " ", "")
		i := index + 1
		if repoExist(repoName) {
			_ = sendMsg(fmt.Sprintf("[%s] Github Trending 总榜单上榜! 名次 %d 当前 star %d", repoName, i, repo.Stars.Count))
		}
	}
}

func syncRepo() {
	for _, language := range Languages {
		getGitHubTrendingByLanguage(language)
	}
	getTotalGitHubTrending()
}

func sendMsg(msg string) error {
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	firends, err := self.Friends()
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	for _, friend := range firends {
		if friend.NickName == FriendName {
			if _, err := self.SendTextToFriend(friend, msg); err != nil {
				fmt.Printf(err.Error())
				return err
			}
		}
	}

	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	Languages = config.Languages
	Repos = config.Repos
	FriendName = config.FriendName
	CronSpec = config.CronSpec

	bot = openwechat.DefaultBot(openwechat.Desktop)
	bot.UUIDCallback = ConsoleQrCode
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()
	if err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		fmt.Printf(err.Error())
		return
	}

	c := cron.New()
	c.AddFunc(CronSpec, func() {
		syncRepo()
	})
	c.Start()

	bot.Block()
}
