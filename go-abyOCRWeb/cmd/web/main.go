package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

/*Полный путь до исполняемого файла finecmd, в ""*/
const ABBYYPath = `"C:\\Program Files (x86)\\ABBYY FineReader 15\\finecmd.exe"`

/*Порт для входящих подключений, c :*/
const Port = ":7000"

/*Папки сохранения исходных и обработанных файлов соответственно*/
const Uploads = "uploads"
const Results = "results"

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

var systemReady bool

func main() {
	_, err := os.Stat(ABBYYPath[1 : len(ABBYYPath)-1])
	systemReady = (err == nil)
	if !systemReady {
		log.Fatal("Не найдена установка FR15")
	}

	addr := flag.String("addr", Port, "Сетевой адрес веб-сервера")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Запуск сервера на http://127.0.0.1%s", *addr)

	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
