package mara

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log *logrus.Logger

func logInit() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
}
