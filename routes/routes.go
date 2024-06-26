package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter configura todas las rutas de la aplicación
func SetupRouter() *gin.Engine {
    router := gin.Default()

    // Configurar el directorio de plantillas
    router.LoadHTMLGlob("templates/*")

    // Ruta inicial con un botón para ingresar
    router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Inicio",
        })
    })

    // Ruta para la pantalla de bienvenida
    router.GET("/bienvenido", func(c *gin.Context) {
        c.HTML(http.StatusOK, "bienvenido.html", gin.H{
            "title": "Bienvenido",
        })
    })

    return router
}