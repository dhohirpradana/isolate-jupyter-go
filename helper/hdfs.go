package helper

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
)

func CheckHDFSConnection() bool {
	env, err := LoadEnv()
	if err != nil {
		return false
	}

	isAccessible := IsURLAccessible(env.HdfsUrl)
	if !isAccessible {
		return false
	}

	return true
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != fiber.StatusOK {
		return errors.New(string(rune(resp.StatusCode)))
	}

	return nil
}

func HDFSRmdir(path string) error {
	env, err := LoadEnv()
	if err != nil {
		fmt.Println("Error rmdir", err.Error())
		return err
	}
	url := env.HdfsUrl + "/webhdfs/v1" + path + "?user.name=hdfs&op=DELETE&recursive=true"

	resp, err := HttpRequest(fiber.MethodDelete, url, nil, nil)
	if err != nil {
		fmt.Println("Error rmdir", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != fiber.StatusOK {
		fmt.Println("Error rmdir", err.Error())
		return errors.New(string(rune(resp.StatusCode)))
	}

	fmt.Println("OK rmdir")
	return nil
}
