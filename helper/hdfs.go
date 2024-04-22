package helper

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func CheckHDFSConnection() bool {
	env, err := LoadEnv()
	if err != nil {
		return false
	}
	resp, err := HttpRequest(fiber.MethodGet, env.HdfsUrl, nil, nil)

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}
