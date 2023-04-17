package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	gen "github.com/zinkeylin/UniqueNumbersGenerator"
)

// структура для хранения параметров
type parameters struct {
	Limits, Threads int
}

var (
	// параметры
	params parameters
	// апгрейдер
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func main() {
	// привязка функций к адресам
	http.HandleFunc("/", root)
	http.HandleFunc("/run", run)
	http.HandleFunc("/ws", ws)

	fmt.Println("Server is listening...")
	fmt.Println("link:http://localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}

func root(w http.ResponseWriter, r *http.Request) {
	// отображение index.html
	http.ServeFile(w, r, "html/index.html")
}

func run(w http.ResponseWriter, r *http.Request) {

	// если к нам пришли данные
	if r.Method == http.MethodPost {

		// читаем параметры
		err := readParams(r.Body, &params)
		if err != nil {
			errorResponce(w, http.StatusBadRequest, err)
		}
	}
}

func ws(w http.ResponseWriter, r *http.Request) {
	// чтение прошло успешно, открываем WEBSocket-соединение
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorResponce(w, http.StatusInternalServerError, err)
		return
	}

	// соединение установлено, кладём закрытие соединения в defer stack
	defer conn.Close()
	// Канал для приёма чисел
	out := make(chan int)

	// запуск хэндлера
	go gen.Handler(params.Limits, params.Threads, out)

	for i := 0; i < params.Limits; i++ {
		// читаем число из канала
		num := <-out
		fmt.Println(num)
		// отправка числа на frontend
		err = conn.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(num)))
		if err != nil {
			errorResponce(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func errorResponce(w http.ResponseWriter, errorResponceCode int, err error) (error) {
	// записываем сообщение об ошибке в респонс
	w.WriteHeader(errorResponceCode)
	_, err = w.Write([]byte("error: " + err.Error()))
	if err != nil {
		return err
	}
	return nil
}

func readParams(reqBody io.ReadCloser, params *parameters) error {
	// читаем тело запроса
	bs, err := ioutil.ReadAll(reqBody)
	if err != nil {
		return err
	}
	// кладём параметры в params
	err = json.Unmarshal(bs, &params)
	if err != nil {
		return err
	}
	// запускаем валидатор
	err = validator(*params)
	if err != nil {
		return err
	}
	return nil
}

func validator(params parameters) error {
	// валидация значений параметров
	if params.Limits <= 0 {
		return errors.New("invalid value for limits")
	}
	if params.Threads <= 0 {
		return errors.New("invalid value for threads")
	}
	return nil
}
