package core

import (
	"github.com/goSc4n/goSc4n/libs"
	"github.com/goSc4n/goSc4n/utils"
	"time"

)

// Background main function to call other background task
func Background(options libs.Options) {
	utils.DebugF("Checking backround task")
	time.Sleep(time.Duration(options.Refresh) * time.Second)

}

