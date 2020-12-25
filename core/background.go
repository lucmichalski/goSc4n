package core

import (
	"github.com/jaeles-project/jaeles/utils"
	"time"

	"github.com/jaeles-project/jaeles/libs"
)

// Background main function to call other background task
func Background(options libs.Options) {
	utils.DebugF("Checking backround task")
	time.Sleep(time.Duration(options.Refresh) * time.Second)

}

