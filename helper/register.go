package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/validator.v2"
	"io"
	"isolate-jupyter-go/entity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type RegisterHandler struct {
}

func InitRegister() RegisterHandler {
	return RegisterHandler{}
}

func (h RegisterHandler) KubeClientTest(c *fiber.Ctx) (err error) {
	client, err := CreateClient("kubeconfig")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	deployList, err := client.AppsV1().Deployments("backend").List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Now you can iterate over deployList.Items to access individual deployments
	for _, deploy := range deployList.Items {
		// Access deployment details
		fmt.Printf("Deployment Name: %s\n", deploy.Name)
	}

	deploy, err := client.AppsV1().Deployments("backend").Get(
		context.Background(),
		"cors-deployment",
		metav1.GetOptions{},
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	fmt.Println(deploy.Status)

	return c.SendString("OK")
}

func (h RegisterHandler) Register(c *fiber.Ctx) (err error) {
	env, err := LoadEnv()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

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

	// HDFS Make Dir
	path := "/usersapujagad/" + register.Company + "/" + register.Username
	err = HDFSMkdir(path)
	if err != nil {
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, "HDFS mkdir not OK")
	}

	// Generate YAML
	yamlPath, err := GenerateYaml(register.Username, port)
	if err != nil {
		_ = HDFSRmdir(path)
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	err = KubeExec("cat", []string{yamlPath})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Kubectl Apply YAML
	applyArgs := []string{"-n", "sapujagad2", "apply", "-f", yamlPath, "--kubeconfig", "kubeconfig"}
	err = KubeExec("kubectl", applyArgs)
	if err != nil {
		_ = HDFSRmdir(path)
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Check is Jupyter Server Ready
	jupyterUrl := env.MasterUrl + ":" + port
	isJupyterReady := IsURLAccessibleRecursive(jupyterUrl, 20, 3*time.Second)
	if !isJupyterReady {
		deleteArgs := []string{"-n", "sapujagad2", "delete", "-f", yamlPath, "--kubeconfig", "kubeconfig"}
		_ = KubeExec("kubectl", deleteArgs)
		_ = HDFSRmdir(path)
		_ = DeleteUser(register.Username)

		return fiber.NewError(fiber.StatusInternalServerError, "Jupyter server not available, deleting services")
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send([]byte("OK"))
}

func (h RegisterHandler) DeleteUser(c *fiber.Ctx) (err error) {
	id := c.Params("id")

	env, err := LoadEnv()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	token, err := getToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	resp, err := HttpRequest(fiber.MethodGet, env.PbUserUrl+"/"+id, nil, headers)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode == fiber.StatusNotFound {
		return fiber.NewError(fiber.StatusNotFound, "User with id "+id+" not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	username, ok := data["username"].(string)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Error get username")
	}

	company, ok := data["company"].(string)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Error get company")
	}

	// Generate YAML
	yamlPath, err := GenerateYaml(username, "00000")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	deleteArgs := []string{"-n", "sapujagad2", "delete", "-f", yamlPath, "--kubeconfig kubeconfig"}
	_ = KubeExec("kubectl", deleteArgs)

	path := "/usersapujagad/" + company + "/" + username
	_ = HDFSRmdir(path)

	_ = DeleteUser(username)

	return c.SendString("User id, " + id + " username, " + username + " company, " + company + " deleted")
}
