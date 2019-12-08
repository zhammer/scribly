package main

import "github.com/gin-gonic/gin"
import "net/http"

func homepage(context *gin.Context) {
	context.HTML(http.StatusOK, "index.html", gin.H{})
}

func main() {
	router := gin.Default()
	router.LoadHTMLFiles("go_templates/index.html")
	router.Static("/static", "./static")

	router.GET("/", homepage)

	router.Run()
}
