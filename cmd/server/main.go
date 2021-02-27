package main

import (
	"github.com/gotosocial/server/internal/api"
)

func main() {
	router := api.NewRouter()
	router.Route()
}
