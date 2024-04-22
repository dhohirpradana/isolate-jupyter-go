package helper

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func CheckHDFSConnection() error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	resp, err := HttpRequest(fiber.MethodGet, env.HdfsUrl, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New("HDFS unknown error")
}

func HDFSMkdir(path string) error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}

	url := env.HdfsUrl + "/webhdfs/v1" + path + "?user.name=hdfs&op=MKDIRS&op=SETPERMISSION&permission=770"

	resp, err := HttpRequest(fiber.MethodPut, url, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		return errors.New(string(rune(resp.StatusCode)))
	}

	return nil
}

func HDFSRmdir(path string) error {
	env, err := LoadEnv()
	if err != nil {
		return err
	}
	url := env.HdfsUrl + "/webhdfs/v1" + path + "?user.name=hdfs&op=DELETE&recursive=true"

	resp, err := HttpRequest(fiber.MethodDelete, url, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		return errors.New(string(rune(resp.StatusCode)))
	}

	return nil
}
