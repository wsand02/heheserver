package utils

import (
	"log"
	"net/http"
)

func HttpLogErr(w http.ResponseWriter, err error, msg string, code int) {
	log.Println(err.Error())
	http.Error(w, msg, code)
}
