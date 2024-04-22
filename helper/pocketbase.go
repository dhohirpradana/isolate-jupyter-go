package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
)

func CheckPBConnection() bool {
	env, err := LoadEnv()
	if err != nil {
		return false
	}
	resp, err := HttpRequest(fiber.MethodGet, env.PbUrl, nil, nil)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return true
	} else {
		return false
	}
}

func getToken() (string, error) {
	env, err := LoadEnv()
	if err != nil {
		return "", err
	}

	identity := env.PbAdminMail
	password := env.PbAdminPassword

	jsonAdminLogin := []byte(`{"identity":"` + identity + `","password":"` + password + `"}`)

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
	}

	resp, err := HttpRequest(fiber.MethodPost, env.PbAdminLoginUrl, bytes.NewReader(jsonAdminLogin), headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return "", err
	}

	token, ok := data["token"].(string)
	if !ok {
		return "", errors.New("token not found or not a string")
	}

	return token, nil
}

func CheckUser(email, username string) error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	token, err := getToken()
	if err != nil {
		return err
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	resp, err := HttpRequest(fiber.MethodGet, env.PbUserUrl, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return err
	}

	users, ok := data["items"].([]interface{})
	if !ok {
		return errors.New("'users' field is not an array")
	}

	emailExists := false
	usernameExists := false

	for _, user := range users {
		if itemMap, ok := user.(map[string]interface{}); ok {
			if emailField, ok := itemMap["email"].(string); ok {
				if emailField == email {
					emailExists = true
					break
				}
			}

			if usernameField, ok := itemMap["username"].(string); ok {
				if usernameField == username {
					usernameExists = true
					break
				}
			}
		}
	}

	if emailExists || usernameExists {
		return errors.New("pocketbase user already exists")
	}

	return nil
}

func CreateUser(email, username, password, firstname, lastname, jPort, createdBy, company string) error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"email":           email,
		"username":        username,
		"password":        password,
		"passwordConfirm": password,
		"firstName":       firstname,
		"lastName":        lastname,
		"role":            "authenticated",
		"jToken":          "test-token-123",
		"jPort":           jPort,
		"createdBy":       createdBy,
		"company":         company,
	}

	token, err := getToken()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	resp, err := HttpRequest(fiber.MethodPost, env.PbUserUrl, bytes.NewReader(jsonBytes), headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
