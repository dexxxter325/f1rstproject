package e

import "fmt"

func Wrap(msg string, err error) error { //msg-текст сообщения с подсказкой
	return fmt.Errorf("%s:%w", msg, err) //ненулевая ошибка
}
func WrapIfErr(msg string, err error) error { //msg-текст сообщения с подсказкой
	if err == nil {
		return nil
	}
	return Wrap(msg, err) //если ненулевая то юзаем 1 функ.
}
