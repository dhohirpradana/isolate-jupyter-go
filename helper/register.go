package helper

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/validator.v2"
	"isolate-jupyter-go/entity"
)

type RegisterHandler struct {
}

func InitRegister() RegisterHandler {
	return RegisterHandler{}
}

func (h RegisterHandler) Register(c *fiber.Ctx) (err error) {
	var register entity.Register

	if err := c.BodyParser(&register); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := validator.Validate(register); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if len(register.Password) < 6 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Password length min 6")
	}

	isUserExists, err := CheckUser(register.Email, register.Username)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if isUserExists {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "User already exists")
	}

	registerBytes, err := json.Marshal(register)
	if err != nil {
		fmt.Println("Error serializing struct:", err)
		return
	}

	port, err := UnusedPort()
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	fmt.Println("UNUSED PORT", port)

	// Generate YAML
	err = GenerateYaml(register.Username, port)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// Kubectl
	//cmd := exec.Command("kubectl", "get", "node", "--kubeconfig", "kubeconfig")
	//
	//cmd.Stdout = os.Stdout
	//
	//if err := cmd.Run(); err != nil {
	//	fmt.Println("could not run command: ", err)
	//}

	// HTTP
	//body := bytes.NewReader([]byte{})
	//
	//headers := map[string]string{
	//	fiber.HeaderContentType: fiber.MIMEApplicationJSON,
	//}

	//resp, err := HttpRequest(fiber.MethodGet, "url", body, headers)
	//if err != nil {
	//	fmt.Printf("Error during HTTP request: %v\n", err)
	//	return
	//}
	//defer resp.Body.Close()

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	err = c.Send(registerBytes)
	if err != nil {
		return err
	}

	return nil
}
