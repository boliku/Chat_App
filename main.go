package main

import (
	"Chat_App/routes"
)

func main() {
    router := routes.SetupRouter()
    router.Run(":8080")
}