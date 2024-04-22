package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
)

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

func CheckUser(email, username string) (bool, error) {
	env, err := LoadEnv()
	if err != nil {
		return false, err
	}

	token, err := getToken()
	if err != nil {
		return false, err
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	resp, err := HttpRequest(fiber.MethodGet, env.PbUserUrl, nil, headers)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return false, err
	}

	users, ok := data["items"].([]interface{})
	if !ok {
		return false, errors.New("'users' field is not an array")
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
		return true, nil
	}

	return false, nil
}
