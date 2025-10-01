package main

import (
	helper "briefcash-jwt/internal/helper/loghelper"

	"github.com/sirupsen/logrus"
)

func main() {
	helper.InitLogger("./briefcash-jwt/resource/app.log", logrus.InfoLevel)
}
