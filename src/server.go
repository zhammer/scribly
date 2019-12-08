package main

import "github.com/gin-gonic/gin"
import "net/http"

func homepage(context *gin.Context) {
	context.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func main() {
	router := gin.Default()
	router.LoadHTMLFiles("templates/index.tmpl")
	router.Static("/static", "./static")

	router.GET("/", homepage)

	router.Run()
}
