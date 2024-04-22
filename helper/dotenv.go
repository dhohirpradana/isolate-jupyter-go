package helper

import (
	"github.com/joho/godotenv"
	"isolate-jupyter-go/entity"
	"os"
)

func LoadEnv() (entity.ENV, error) {
	var env entity.ENV
	err := godotenv.Load()
	if err != nil {
		return env, err
	}

	env.HdfsUrl = os.Getenv("HDFS_URL")
	env.PbUrl = os.Getenv("PB_URL")
	env.PbAdminLoginUrl = os.Getenv("PB_ADMIN_LOGIN_URL")
	env.PbAdminMail = os.Getenv("PB_ADMIN_MAIL")
	env.PbUserUrl = os.Getenv("PB_USER_URL")
	env.PbAdminPassword = os.Getenv("PB_ADMIN_PASSWORD")
	env.Secret = os.Getenv("SECRET")
	env.SupersetApiUrl = os.Getenv("SUPERSET_API_URL")

	return env, nil
}
