package utils

import (
	"net/http"
)

func GetHttpClient() *http.Client {
	return &http.Client{}
}
