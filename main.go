package main

import (
	"github.com/vulcanize/eth-header-sync/cmd"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	cmd.Execute()
}
