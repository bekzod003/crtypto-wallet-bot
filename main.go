package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

type wallet map[string]float64

var db = map[int]wallet{}

const TOKEN = "YOUR_TOKEN"

func main() {
	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		command := strings.Split(update.Message.Text, " ")
		fmt.Println(command)
		userId := update.Message.From.ID

		switch command[0] {
		case "add":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное количество аргументов"))
				continue
			}

			coinAmount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				fmt.Print(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда неверно введена"))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}

			db[userId][command[1]] += coinAmount

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда add успешно выполнена"))
		case "sub":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное количество аргументов"))
				continue
			}

			coinAmount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				fmt.Print(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда неверно введена"))
				continue
			}

			if _, ok := db[userId]; !ok {
				msg := fmt.Sprintf("Криптовалюты %s не было в кошельке", command[1])
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			}

			db[userId][command[1]] -= coinAmount

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда sub усешно выполнена"))
		case "del":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверное количество аргументов"))
				continue
			}

			_, exists := db[userId][command[1]]
			if !exists {
				msg := fmt.Sprintf("Криптовалюты %s не было в кошельке", command[1])
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			}
			delete(db[userId], command[1])

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Команда del успешно выполнена"))
		case "show":
			fmt.Println(db)
			if len(db[userId]) == 0 {
				msg := "Кошелек пуст"
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			}

			dbInfo := ""
			for coinName, amount := range db[userId] {
				usdPrice, err := getUsd(coinName)
				if err != nil {
					fmt.Print(err)
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}

				rubPrice, _ := getRub(coinName)

				dbInfo += fmt.Sprintf("%s: %.2f\n%s в долларах: %.2f\n%s в рублях: %.2f\n\n", strings.ToUpper(coinName), amount, strings.ToUpper(coinName), amount*usdPrice, strings.ToUpper(coinName), amount*rubPrice)
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, dbInfo))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	}

}

func getUsd(currency string) (float64, error) { //usd, err
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", strings.ToUpper(currency))
	resp, err := http.Get(url)

	if err != nil {
		return 0.0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)

	if bRes.Symbol == "" || err != nil {
		return 0.0, errors.New("неверная валюта")
	}

	return bRes.Price, nil
}

func getRub(currency string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sRUB", strings.ToUpper(currency))
	resp, err := http.Get(url)

	if err != nil {
		return 0.0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)

	if bRes.Symbol == "" || err != nil {
		return 0.0, errors.New("неверная валюта")
	}

	return bRes.Price, nil
}
