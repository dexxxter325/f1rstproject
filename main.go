package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"

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
		if update.Message == nil { // If we got a message
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

	myStorage := &MyStorage{}

	isExists, err := myStorage.IsExists(*page) //существует ли ссылка уже
	if err != nil {
		return err
	}

	if isExists {
		return SendMessage(chatID, msgAlreadyExists)
	}

	// Получение имени файла из URL-адреса
	// Получение имени файла из URL-адреса
	filename := path.Base(page.URL)

	// Вычисление хеш-суммы имени файла
	checksum := md5.Sum([]byte(filename))
	checksumStr := hex.EncodeToString(checksum[:])

	if err := myStorage.Save(checksumStr); err != nil {
		return err
	}

	if err := SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

type UserPage struct {
	Username string
	Page     *storage.Page
}

type MyStorage struct {
	Dataa     []string
	offset    int //используется, чтобы получить обновления, начиная не с самого первого, а с некоторого определённого ID.
	storage   storage.Storage
	Data      string
	URL       string
	UserPages []*UserPage
	XData     map[string]string
}

func (s *MyStorage) Save(filename string) error {
	// Реализация метода Save
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Ошибка при закрытии файла:", err)
		}
	}(file)

	_, err = file.WriteString(s.Data)
	if err != nil {
		return err
	}

	fmt.Println("Данные успешно сохранены в файл", filename)
	return nil
}

func (s *MyStorage) IsExists(p storage.Page) (bool, error) {
	// Реализация метода IsExists
	// Проверьте, существует ли страница в хранилище
	// Верните true, если страница существует, и false, если она не существует
	// Верните ошибку, если произошла ошибка при проверке
	for _, data := range s.Dataa {
		if data == p.URL {
			return true, nil
		}
	}
	return false, nil
}

func sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("cant do command:cant send random", err) }()
	myStorage := &MyStorage{}

	page, err := myStorage.PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) { //errors.is позволяет определить, является ли конкретная ошибка заданным типом ошибки(когда нету сохр.стр)
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return SendMessage(chatID, msgNoSavedPages) //если ниче не сохранил
	}
	if err := SendMessage(chatID, page.URL); err != nil { //если удалось найти ссылку
		return err
	}

	err = myStorage.Remove(page) //удаляем ссылку
	if err != nil {
		return err
	}

	return nil
}

func (s *MyStorage) Remove(p *storage.Page) error {
	// Реализация метода Remove
	// Найдите страницу в хранилище и удалите ее
	// Верните ошибку, если страница не найдена или произошла ошибка при удалении
	index := -1
	for i, data := range s.Dataa {
		if data == p.URL {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("page not found: %s", p.URL)
	}
	// Удаление страницы из хранилища
	s.Dataa = append(s.Dataa[:index], s.Dataa[index+1:]...)
	return nil
}

func (s *MyStorage) PickRandom(username string) (*storage.Page, error) {
	// Получить список страниц, связанных с указанным пользователем
	userPages, err := s.getUserPages(username)
	if err != nil {
		return nil, err
	}

	// Проверить, есть ли страницы для выбора случайной страницы
	if len(userPages) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	// Сгенерировать случайный индекс в пределах длины списка страниц
	randomIndex := rand.Intn(len(userPages))

	// Вернуть случайно выбранную страницу
	return userPages[randomIndex], nil
}
func (s *MyStorage) getUserPages(username string) ([]*storage.Page, error) {
	var userPages []*storage.Page

	// Перебираем все страницы пользователей в нашем хранилище
	for _, userPage := range s.UserPages {
		if userPage.Username == username {
			// Если имя пользователся совпадает, добавляем страницу в список
			userPages = append(userPages, userPage.Page)
		}
	}

	// Проверяем, есть ли страницы для выбранного пользователя
	if len(userPages) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	// Возвращаем список страниц пользователя
	return userPages, nil
}

func SendMessage(chatID int, message string) error { //chatID ,чтобы уточнить,куда конкретно отпр.сообщ
	bot, err := tgbotapi.NewBotAPI("6342619263:AAHm5ZpmMEn9ozRabHN4Es3YzzLt_ffocP8")
	msg := tgbotapi.NewMessage(int64(chatID), message)

	_, err = bot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
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
	u, err := url.Parse(text)                                                        //распарсить URL-проанализировать текстовую запись URL и извлечь из него основные компоненты: хост, порт, путь, параметры и фрагмент
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https") //если при разборе нет ошибки и хост из разобарнного url не пустой
}
