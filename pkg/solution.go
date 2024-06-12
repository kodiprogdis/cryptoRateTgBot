package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CoinInfo struct {
	Time         string  `json:"time"`
	AssetIdBase  string  `json:"base"`
	AssetIdQuote string  `json:"quote"`
	Rate         float64 `json:"rate"`
}

func getPrice(amount int, coin1 string, coin2 string) string {
	client := &http.Client{}
	url_coinapi := "https://rest.coinapi.io/v1/exchangerate/" + coin1 + "/" + coin2
	req, err := http.NewRequest("GET", url_coinapi, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CoinAPI-Key", os.Getenv("COINAPI_KEY"))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	var coininfo CoinInfo
	err2 := json.Unmarshal([]byte(respBody), &coininfo)
	if err2 != nil {
		fmt.Println("Ошибка чтения JSON-данных:", err2)
	}
	log.Println(string(respBody))
	return fmt.Sprint(coininfo.Rate * float64(amount))
}

func main() {
	er := godotenv.Load("settings.env")
	if er != nil {
		log.Fatalf("Some error occured. Err: %s", er)
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN")) // create new bot
	if err != nil {
		panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.InlineQuery == nil { // if no inline query, ignore it
			continue
		}
		_l := strings.Split(update.InlineQuery.Query, " ")

		if len(_l) == 3 {
			amount, _ := strconv.Atoi(_l[0])
			title_article := fmt.Sprintf("Неверный формат данных\n @%s 1 USD RUB", bot.Self.UserName)
			article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID,
				title_article,
				getPrice(amount, _l[1], _l[2]))
			article.Description = getPrice(amount, _l[1], _l[2])

			inlineConf := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				CacheTime:     0,
				Results:       []interface{}{article},
			}

			if _, err := bot.Request(inlineConf); err != nil {
				log.Println(err)
			}
		}

	}

}
