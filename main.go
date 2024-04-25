package main

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	RegisterHandler "isolate-jupyter-go/helper"
)

func main() {
	isolateJupyter := RegisterHandler.InitRegister()

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(
		helmet.New(),
	)

	app.Use(cors.New())

	app.Post("/user", isolateJupyter.Register)
	app.Get("/kube-test", isolateJupyter.KubeClientTest)
	app.Post("/test", isolateJupyter.Test)
	app.Delete("/user/:id", isolateJupyter.DeleteUser)
	app.Get("/metrics", monitor.New())

	log.Fatal(app.Listen(":9090"))
}
