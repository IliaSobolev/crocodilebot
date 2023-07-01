package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"
)

type JSONDATA struct {
	UserId string `json:"userid"`
	WordId string `json:"wordid"`
}

func RandomWord(client *redis.Client, ctx context.Context, chat int64, user int64, jsonData []byte) (string, error) {
	var data JSONDATA
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return "", err
	}

	file, err := os.Open("output.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, " ", "")
		lines = append(lines, line)
	}

	if len(lines) > 0 {
		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(lines))
		randomLine := lines[randomIndex]
		data.WordId = randomLine
		if data.UserId == "" {
			data.UserId = strconv.FormatInt(user, 10)
		}

		encodedData, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		err = client.Set(ctx, strconv.FormatInt(chat, 10), encodedData, 5*time.Minute).Err()
		if err != nil {
			return "", err
		}
	}

	return data.WordId, err
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	jsonData, err := ioutil.ReadFile("id.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	b, err := tele.NewBot(pref)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	var (
		selector = &tele.ReplyMarkup{}

		this_word = selector.Data("Показать Слово", "prev")
		new_word  = selector.Data("Новое слово", "next")
	)

	selector.Inline(
		selector.Row(this_word),
		selector.Row(new_word),
	)

	b.Handle(tele.OnText, func(c tele.Context) error {
		if c.Chat().Type == tele.ChatPrivate {
			return nil
		} else {
			userText := strings.ToLower(c.Text())
			exists, err := client.Exists(ctx, strconv.FormatInt(c.Chat().ID, 10)).Result()
			val, err := client.Get(ctx, strconv.FormatInt(c.Chat().ID, 10)).Result()
			if err != nil {
				return nil
			}
			var data JSONDATA
			err = json.Unmarshal([]byte(val), &data)
			if err != nil {
				return (err)
			}

			if exists == 1 && userText == data.WordId && strconv.FormatInt(c.Sender().ID, 10) != data.UserId {
				err := client.Del(ctx, strconv.FormatInt(c.Chat().ID, 10)).Err()
				if err != nil {
					return (err)
				} else {
					return c.Send("Ты победил!")
				}
			} else if userText == data.WordId && strconv.FormatInt(c.Sender().ID, 10) == data.UserId {
				err := client.Del(ctx, strconv.FormatInt(c.Chat().ID, 10)).Err()
				if err != nil {
					return err
				}
				return c.Send("Так нельзя!")
			}
		}
		return nil
	})
	b.Handle("/start", func(c tele.Context) error {
		chat := c.Chat().ID
		chatT := c.Chat().Type
		user := c.Sender().ID

		if chatT == tele.ChatPrivate {
			return c.Send("Для того чтобы начать игру, добавьте бота в группу")
		} else {
			exists, err := client.Exists(ctx, strconv.FormatInt(chat, 10)).Result()
			if err != nil {
				return err
			}

			if exists == 1 {
				return c.Send("Игра уже началась. Ожидайте 5 минут")
			} else {
				_, err := RandomWord(client, ctx, chat, user, jsonData)
				if err != nil {
					return err
				}
				return c.Send("Угадай слово", selector)
			}
		}
	})

	b.Handle(&this_word, func(c tele.Context) error {
		chat := c.Chat().ID
		user := c.Sender().ID
		val, err := client.Get(ctx, strconv.FormatInt(chat, 10)).Result()
		if err != nil {
			return err
		}
		var data JSONDATA
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			return (err)
		}

		if data.UserId == strconv.FormatInt(user, 10) {
			return c.Respond(&tele.CallbackResponse{
				Text:      fmt.Sprintf("Вы загадываете слово: %s", data.WordId),
				ShowAlert: true,
			})
		} else {
			return c.Respond(&tele.CallbackResponse{
				Text:      "Сейчас не вы загадываете слово!",
				ShowAlert: true,
			})
		}

	})

	b.Handle(&new_word, func(c tele.Context) error {
		chat := c.Chat().ID
		user := c.Sender().ID
		val, err := client.Get(ctx, strconv.FormatInt(chat, 10)).Result()
		if err != nil {
			return err
		}
		var data JSONDATA
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			return (err)
		}

		if data.UserId == strconv.FormatInt(user, 10) {
			word, err := RandomWord(client, ctx, chat, user, jsonData)
			if err != nil {
				return (err)
			}
			return c.Respond(&tele.CallbackResponse{
				Text:      fmt.Sprintf("Новое слово: %s", word),
				ShowAlert: true,
			})
		} else {
			return c.Respond(&tele.CallbackResponse{
				Text:      "Сейчас не вы загадываете слово!",
				ShowAlert: true,
			})
		}

	})

	b.Start()
}
