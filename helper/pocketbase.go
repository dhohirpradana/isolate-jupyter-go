package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
)

func CheckPBConnection() bool {
	env, err := LoadEnv()
	if err != nil {
		return false
	}

	isAccessible := IsURLAccessible(env.PbUrl)
	if !isAccessible {
		return false
	}

	return true
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	token, ok := data["token"].(string)
	if !ok {
		return "", errors.New("token not found or not a string")
	}

	return token, nil
}

func ListUsers() ([]interface{}, error) {
	env, err := LoadEnv()
	if err != nil {
		return nil, err
	}

	token, err := getToken()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	resp, err := HttpRequest(fiber.MethodGet, env.PbUserUrl, nil, headers)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	users, ok := data["items"].([]interface{})
	if !ok {
		return nil, errors.New("'users' field is not an array")
	}

	return users, nil
}

func CheckUser(username, email string) error {
	users, err := ListUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if itemMap, ok := user.(map[string]interface{}); ok {
			if emailField, ok := itemMap["email"].(string); ok {
				if emailField == email {
					return errors.New("pocketbase user already exists")
				}
			}

			if usernameField, ok := itemMap["username"].(string); ok {
				if usernameField == username {
					return errors.New("pocketbase user already exists")
				}
			}
		}
	}

	return nil
}

func CheckUserById(id string) error {
	users, err := ListUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if itemMap, ok := user.(map[string]interface{}); ok {
			if idField, ok := itemMap["id"].(string); ok {
				if idField == id {
					return nil
				}
			}
		}
	}

	return errors.New("pocketbase user for create by not exists")
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return nil
}

func DeleteUser(username string) error {
	env, err := LoadEnv()
	if err != nil {
		fmt.Println("Error delete user", err.Error())
		return err
	}

	token, err := getToken()
	if err != nil {
		fmt.Println("Error delete user", err.Error())
		return err
	}

	headers := map[string]string{
		fiber.HeaderContentType: fiber.MIMEApplicationJSON,
		"Authorization":         "Bearer " + token,
	}

	users, err := ListUsers()
	if err != nil {
		fmt.Println("Error delete user", err.Error())
		return err
	}

	for _, user := range users {
		if itemMap, ok := user.(map[string]interface{}); ok {
			if usernameField, ok := itemMap["username"].(string); ok {
				if usernameField == username {
					userId := itemMap["id"].(string)
					resp, err := HttpRequest(fiber.MethodDelete, env.PbUserUrl+"/"+userId, nil, headers)
					if err != nil {
						fmt.Println("Error delete user", err.Error())
						return err
					}
					func(Body io.ReadCloser) {
						_ = Body.Close()
					}(resp.Body)

					fmt.Println("OK delete user")
					return nil
				}
			}
		}
	}

	fmt.Println("OK delete user")
	return nil
}
