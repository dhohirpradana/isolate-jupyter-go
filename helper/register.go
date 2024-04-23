package helper

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/validator.v2"
	"isolate-jupyter-go/entity"
	"os"
	"os/exec"
)

type RegisterHandler struct {
}

func InitRegister() RegisterHandler {
	return RegisterHandler{}
}

func (h RegisterHandler) Register(c *fiber.Ctx) (err error) {
	okPB := CheckPBConnection()
	if !okPB {
		return fiber.NewError(fiber.StatusInternalServerError, "Pocketbase service not OK")
	}

	okHdfs := CheckHDFSConnection()
	if !okHdfs {
		return fiber.NewError(fiber.StatusInternalServerError, "HDFS service not OK")
	}

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

	// Check User
	err = CheckUser(register.Username, register.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// Check CreatedBy
	err = CheckUserById(register.CreatedBy)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// Get Unused Port
	port, err := UnusedPort()
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	fmt.Println("UNUSED PORT", port)

	// Create User Pocketbase
	err = CreateUser(
		register.Email,
		register.Username,
		register.Password,
		register.FirstName,
		register.LastName,
		port,
		register.CreatedBy,
		register.Company,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Pocketbase create user not OK")
	}

	// HDFS
	path := "/usersapujagad/" + register.Company + "/" + register.Username
	err = HDFSMkdir(path)
	if err != nil {
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, "HDFS mkdir not OK")
	}

	// Generate YAML
	err = GenerateYaml(register.Username, port)
	if err != nil {
		_ = HDFSRmdir(path)
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Kubectl
	cmd := exec.Command("kubectl", "get", "node", "--kubeconfig", "kubeconfig")

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		_ = DeleteFile(register.Username)
		_ = HDFSRmdir(path)
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, "cannot execute command: "+err.Error())
	}

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
	err = c.Send([]byte("OK"))
	if err != nil {
		return err
	}

	return nil
}
