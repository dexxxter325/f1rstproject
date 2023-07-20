package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"tgbot/lib/e"
)

type Storage interface {
	Save(p *Page) error                        //Метод для сохранения страницы в хранилище
	PickRandom(userName string) (*Page, error) //Метод для получения случайной страницы для указанного пользователя
	Remove(p *Page) error                      //Удаляет страницу из хранилища
	IsExists(p *Page) (bool, error)            //сообщает,существует ли страница
}

var ErrNoSavedPages = errors.New("no saved pages") // errors.New() позволяет легко создавать новые ошибки с указанным текстом
type Page struct {                                 //page-основной тип данных,с которым будет работать storage.Это страница,на которую бот заходит,передя по ссылке,кот. мы скинули
	URL      string
	UserName string
}

func (p Page) Hash() (string, error) { //Hash-процесс преобразования данных в фиксированную длину цифровую "отметку", называемую хэшем.
	h := sha1.New()                                     //sha1-способ работы с хешами,New создает и возвращает объект Hash, который можно использовать для вычисления SHA1 хэшей
	if _, err := io.WriteString(h, p.URL); err != nil { //io.WriteString-функция для записи строки в объекты, которые умеют записывать байты.
		return "", e.Wrap("cant calculate hash", err) //"" чтобы вернуть "пустой" результат в случае ошибки, а не некорректный хеш
	}
	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("cant calculate hash", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil //sum-итоговый хэш,sprintf преобразует в строку
}
