package main

import (
	"Chat_App/routes"
	"Chat_App/utils"
)

func main() {
    router := routes.SetupRouter()

    utils.ConnectDatabase()

    router.Run(":8080")
}