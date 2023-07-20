package main

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"
	"tgbot/lib/e"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const msgHelp = `I can save and keep you pages. Also I can offer you them to read.

In order to save the page, just send me al link to it.

In order to get a random page from your list, send me command /rnd.
Caution! After that, this page will be removed from your list!С уважением, Petr Tate`
const msgHello = "wassup man🙃 \n\n заходи не бойся ,выходи не плачь\n\n" + msgHelp
const (
	msgUnknownCommand = "это фиаско🤷‍♂️"
	msgNoSavedPages   = "👀у вас нету сохраненных ссылок;(((("
	msgSaved          = "успешно сохранено,сэр👌"
	msgAlreadyExists  = "босс,вы же уже сохраняли эту ссылку👉👈"
)
const (
	RndCmd   = "/rnd" //рандомная ссылка
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("6342619263:AAHm5ZpmMEn9ozRabHN4Es3YzzLt_ffocP8")
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			continue

		}
		go handleCommand(bot, update.Message)
	}
}
func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	username := message.From.UserName
	text := message.Text

	text = strings.TrimSpace(text)
	log.Printf("Got new command '%s' from '%s'", text, username)

	err := doCmd(text, int(chatID), username)
	if err != nil {
		log.Println(err)
	}
}
func doCmd(text string, chatID int, username string) error { //doCmd-api роутера(смотрит на отправленное сообщение и понимает,какую команду выполнить)
	text = strings.TrimSpace(text) //удаляем пробелы из текста
	log.Printf("got new command '%s'from '%s", text, username)
	if isAddCmd(text) { //если отпр.сообщ.явл ссылкой
		return savePage(chatID, text, username)
	}
	switch text {
	case RndCmd:
		return sendRandom(chatID, username)
	case HelpCmd:
		return sendHelp(chatID)
	case StartCmd:
		return sendHello(chatID)
	default:
		return SendMessage(chatID, msgUnknownCommand)

	}

}
func savePage(chatID int, pageURL string, username string) (err error) { //сохранение страницы в хранилище
	defer func() { err = e.WrapIfErr("cant do command:save page", err) }() //Она обрабатывает ошибку только если она не nil.
	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}
	isExists, err := IsExists() //существует ли ссылка уже
	if err != nil {
		return err
	}
	if isExists {
		return SendMessage(chatID, msgAlreadyExists)
	}
	if err := Save(page); err != nil {
		return err
	}
	if err := SendMessage(chatID, msgSaved); err != nil {
		return err
	}
	return nil
}

func Save(page *storage.Page) error {
	return nil

}
func IsExists() (bool, error) {
	return IzExists(msgAlreadyExists), nil

}

func IzExists(text string) bool {
	q := url.Values{}
	q.Add("text", text)

	return true

}

func sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("cant do command:cant send random", err) }()
	page, err := PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) { //errors.is позволяет определить, является ли конкретная ошибка заданным типом ошибки(когда нету сохр.стр)
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return SendMessage(chatID, msgNoSavedPages) //если ниче не сохранил
	}
	if err := SendMessage(chatID, page.URL); err != nil { //если удалось найти ссылку
		return err
	}
	return Remove(page) //удаляем ссылку

}

func Remove(page *storage.Page) error {
	return nil

}

func PickRandom(username string) (*storage.Page, error) {
	return nil, nil

}
func SendMessage(chatID int, text string) error { //chatID ,чтобы уточнить,куда конкретно отпр.сообщ
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	return nil
}

func sendHello(chatID int) error {
	return SendMessage(chatID, msgHello)
}
func sendHelp(chatID int) error {
	return SendMessage(chatID, msgHelp)
}
func isAddCmd(text string) bool {
	return isURL(text)

}
func isURL(text string) bool {
	u, err := url.Parse(text)         //распарсить URL-проанализировать текстовую запись URL и извлечь из него основные компоненты: хост, порт, путь, параметры и фрагмент
	return err == nil && u.Host != "" //если при разборе нет ошибки и хост из разобарнного url не пустой
}
