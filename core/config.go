package core

import (
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/goSc4n/goSc4n/libs"
	"github.com/goSc4n/goSc4n/utils"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"os"
	"path"
	_ "path/filepath"
)


func UpdateSignature(options libs.Options) {
	signPath := path.Join(options.RootFolder, "base-signatures")
	url := libs.SIGNREPO

	if options.Config.Repo != "" {
		url = options.Config.Repo
	}

	utils.GoodF("Cloning Signature from: %v", url)
	if utils.FolderExists(signPath) {
		utils.InforF("Remove: %v", signPath)
		os.RemoveAll(signPath)
		os.RemoveAll(options.ResourcesFolder)
		os.RemoveAll(options.ThirdPartyFolder)
	}
	if options.Config.PrivateKey != "" {
		cmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git clone --depth=1 %v %v", options.Config.PrivateKey, url, signPath)
		Execution(cmd)
	} else {
		var err error
		if options.Server.Username != "" && options.Server.Password != "" {
			_, err = git.PlainClone(signPath, false, &git.CloneOptions{
				Auth: &http.BasicAuth{
					Username: options.Config.Username,
					Password: options.Config.Password,
				},
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Depth:             1,
				Progress:          os.Stdout,
			})
		} else {
			_, err = git.PlainClone(signPath, false, &git.CloneOptions{
				URL:               url,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Depth:             1,
				Progress:          os.Stdout,
			})
		}

		if err != nil {
			utils.ErrorF("Error to clone Signature repo: %v - %v", url, err)
			return
		}
	}
}

// InitConfig init config
func InitConfig(options *libs.Options) {
	options.RootFolder = utils.NormalizePath(options.RootFolder)
	options.Server.DBPath = path.Join(options.RootFolder, "sqlite3.db")
	// init new root folder
	if !utils.FolderExists(options.RootFolder) {
		utils.InforF("Init new config at %v", options.RootFolder)
		os.MkdirAll(options.RootFolder, 0750)
		// cloning default repo
		//UpdatePlugins(*options)
		UpdateSignature(*options)
	}

	configPath := path.Join(options.RootFolder, "config.yaml")
	v := viper.New()
	v.AddConfigPath(options.RootFolder)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if !utils.FileExists(configPath) {
		utils.InforF("Write new config to: %v", configPath)
		// save default config if not exist
		bind := "http://127.0.0.1:5000"
		v.SetDefault("defaultSign", "*")
		v.SetDefault("cors", "*")
		// default credential
		v.SetDefault("username", "goSc4n")
		v.SetDefault("password", utils.GenHash(utils.GetTS())[:10])
		v.SetDefault("secret", utils.GenHash(utils.GetTS()))
		v.SetDefault("bind", bind)
		v.WriteConfigAs(configPath)

	} else {
		if options.Debug {
			utils.InforF("Load config from: %v", configPath)
		}
		b, _ := ioutil.ReadFile(configPath)
		v.ReadConfig(bytes.NewBuffer(b))
	}

	// WARNING: change me if you really want to deploy on remote server
	// allow all origin
	options.Server.Cors = v.GetString("cors")
	options.Server.JWTSecret = v.GetString("secret")
	options.Server.Username = v.GetString("username")
	options.Server.Password = v.GetString("password")

	// store default credentials for Burp plugin
	burpConfigPath := path.Join(options.RootFolder, "burp.json")
	if !utils.FileExists(burpConfigPath) {
		jsonObj := gabs.New()
		jsonObj.Set("", "JWT")
		jsonObj.Set(v.GetString("username"), "username")
		jsonObj.Set(v.GetString("password"), "password")
		bind := v.GetString("bind")
		if bind == "" {
			bind = "http://127.0.0.1:5000"
		}
		jsonObj.Set(fmt.Sprintf("http://%v/api/parse", bind), "endpoint")
		utils.WriteToFile(burpConfigPath, jsonObj.String())
		if options.Verbose {
			utils.InforF("Store default credentials for client at: %v", burpConfigPath)
		}
	}

	// set some default config
	options.ResourcesFolder = path.Join(utils.NormalizePath(options.RootFolder), "resources")
	options.ThirdPartyFolder = path.Join(utils.NormalizePath(options.RootFolder), "thirdparty")

	// create output folder
	var err error
	err = os.MkdirAll(options.Output, 0750)
	if err != nil && options.NoOutput == false {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %s -- %s\n", err, options.Output)
		os.Exit(1)
	}
	if options.SummaryOutput == "" {
		options.SummaryOutput = path.Join(options.Output, "goSc4n-summary.txt")
	}
	if options.SummaryVuln == "" {
		options.SummaryVuln = path.Join(options.Output, "vuln-summary.txt")
	}


	dbSize := utils.GetFileSize(options.Server.DBPath)
	if dbSize > 5.0 {
		utils.WarningF("Your Database size look very big: %vGB", fmt.Sprintf("%.2f", dbSize))
		utils.WarningF("Consider clean your db with this command: 'goSc4n config -a clear' or just remove your '~/.goSc4n/'")
	}
	utils.InforF("Summary output: %v", options.SummaryOutput)

	if options.ChunkRun {
		if options.ChunkDir == "" {
			options.ChunkDir = path.Join(os.TempDir(), "goSc4n-chunk-data")
		}
		os.MkdirAll(options.ChunkDir, 0755)
	}

}
