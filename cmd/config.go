package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/goSc4n/goSc4n/tree/hoangnm/database"
	"github.com/goSc4n/goSc4n/tree/hoangnm/libs"
	"github.com/goSc4n/goSc4n/tree/hoangnm/utils"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Configuration CLI",
		Long:  libs.Banner(),
		RunE:  runConfig,

	}
	configCmd.SetHelpFunc(configHelp)
	RootCmd.AddCommand(configCmd)

}

func runConfig(cmd *cobra.Command, args []string) error {
	sort.Strings(args)
	// print more help
	helps, _ := cmd.Flags().GetBool("hh")
	if helps == true {
		HelpMessage()
		os.Exit(1)
	}
	// turn on verbose by default
	options.Verbose = true
	polling, _ := cmd.Flags().GetBool("poll")
	// polling all oob
	if polling == true {
		secret, _ := cmd.Flags().GetString("secret")
		collabFile, _ := cmd.Flags().GetString("collab")
		collabs := utils.ReadingLines(collabFile)
		for _, collab := range collabs {
			database.ImportCollab(secret, collab)
		}
	}

	action, _ := cmd.Flags().GetString("action")
	// backward compatible
	if action == "" && len(args) > 0 {
		action = args[0]
	}
	getEnv(&options)

	switch action {
	case "init":
		if options.Config.Forced {
			os.RemoveAll(options.SignFolder)
		}
		reloadSignature(options.SignFolder, options.Config.SkipMics)
		break
	case "clear":
		utils.GoodF("Cleaning your DB")
		database.CleanScans()
		database.CleanSigns()
		database.CleanRecords()
		break
	case "clean":
		utils.InforF("Cleaning root folder: %v", options.RootFolder)
		os.RemoveAll(options.RootFolder)
		break
	case "cred":
		database.CreateUser(options.Config.Username, options.Config.Password)
		utils.GoodF("Create new credentials %v:%v \n", options.Config.Username, options.Config.Password)
		break
	case "oob":
		secret, _ := cmd.Flags().GetString("secret")
		collabFile, _ := cmd.Flags().GetString("collab")
		collabs := utils.ReadingLines(collabFile)
		for _, collab := range collabs {
			database.ImportCollab(secret, collab)
		}
		break
	case "reload":
		os.RemoveAll(path.Join(options.RootFolder, "base-signatures"))
		InitDB()
		reloadSignature(options.SignFolder, options.Config.SkipMics)
		break
	case "add":
		addSignature(options.SignFolder)
		break
	case "select":
		SelectSign()
		if len(options.SelectedSigns) == 0 {
			fmt.Fprintf(os.Stderr, "[Error] No signature loaded\n")
			fmt.Fprintf(os.Stderr, "Use 'goSc4n -h' for more information about a command.\n")
		} else {
			utils.GoodF("Signatures Loaded: %v", strings.Join(options.SelectedSigns, " "))
		}
		break
	default:
		HelpMessage()
	}
	CleanOutput()
	return nil
}

// addSignature add active signatures from a folder
func addSignature(signFolder string) {
	signFolder = utils.NormalizePath(signFolder)
	if !utils.FolderExists(signFolder) {
		utils.ErrorF("Signature folder not found: %v", signFolder)
		return
	}
	allSigns := utils.GetFileNames(signFolder, ".yaml")
	if allSigns != nil {
		utils.InforF("Add Signature from: %v", signFolder)
		for _, signFile := range allSigns {
			database.ImportSign(signFile)
		}
	}
}

// reloadSignature signature
func reloadSignature(signFolder string, skipMics bool) {
	signFolder = utils.NormalizePath(signFolder)
	if !utils.FolderExists(signFolder) {
		utils.ErrorF("Signature folder not found: %v", signFolder)
		return
	}
	utils.GoodF("Reload signature in: %v", signFolder)
	database.CleanSigns()
	SignFolder, _ := filepath.Abs(path.Join(options.RootFolder, "base-signatures"))
	if signFolder != "" && utils.FolderExists(signFolder) {
		SignFolder = signFolder
	}
	allSigns := utils.GetFileNames(SignFolder, ".yaml")
	if len(allSigns) > 0 {
		utils.InforF("Load Signature from: %v", SignFolder)
		for _, signFile := range allSigns {
			if skipMics {
				if strings.Contains(signFile, "/mics/") {
					utils.DebugF("Skip sign: %v", signFile)
					continue
				}

				if strings.Contains(signFile, "/exper/") {
					utils.DebugF("Skip sign: %v", signFile)
					continue
				}
			}
			utils.DebugF("Importing signature: %v", signFile)
			err := database.ImportSign(signFile)
			if err != nil {
				utils.ErrorF("Error importing signature: %v", signFile)
			}
		}
	}

	signPath := path.Join(options.RootFolder, "base-signatures")
	resourcesPath := path.Join(signPath, "resources")
	thirdpartyPath := path.Join(signPath, "thirdparty")

	// copy it to base signature folder
	if !utils.FolderExists(signPath) {
		utils.CopyDir(signFolder, signPath)
	}


	if utils.FolderExists(resourcesPath) {
		utils.MoveFolder(resourcesPath, options.ResourcesFolder)
	}
	if utils.FolderExists(thirdpartyPath) {
		utils.MoveFolder(thirdpartyPath, options.ThirdPartyFolder)
	}

}

func configHelp(_ *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	HelpMessage()
}

func rootHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	helps, _ := cmd.Flags().GetBool("hh")
	if helps {
		fmt.Println(cmd.UsageString())
		return
	}
	RootMessage()
}

// RootMessage print help message
func RootMessage() {
	h := "\nUsage:\n goSc4n scan|server|config|fuzz|spider [options]\n"
	h += " goSc4n scan|server|config|report|fuzz|spider -h -- Show usage message\n"
	h += "\nSubcommands:\n"
	h += "  goSc4n scan   --  Scan list of URLs based on selected signatures\n"
	h += "  goSc4n server --  Start API server\n"
	h += "  goSc4n config --  Configuration CLI \n"
	h += "  goSc4n report --  Generate HTML report based on scanned output \n"
	h += "  goSc4n fuzz   --  fuzzing one or many sites \n"
	h += "  goSc4n spider --  crawler one or many sites \n"
//	h += `
//Core Flags:
//  -c, --concurrency int         Set the concurrency level (default 20)
//  -o, --output string           Output folder name (default "out")
//  -s, --signs strings           Signature selector (Multiple -s flags are accepted)
//  -L, --level int               Filter signatures by level (default 1)
//  -p, --params strings          Custom params -p='foo=bar' (Multiple -p flags are accepted)
//  -H, --headers strings         Custom headers (e.g: -H 'Referer: {{.BaseURL}}') (Multiple -H flags are accepted)
//
//Mics Flags:
//      --proxy string            proxy
//      --timeout int             HTTP timeout (default 20)
//      --debug                   Debug
//  -v, --verbose                 Verbose
//      --no-db                   Disable Database
//  -S, --selectorFile string     Signature selector from file
//  -J, --format-input            Enable special input format (default is false)
//  -f, --found string            Run host OS command when vulnerable found
//  -O, --summaryOutput string    Summary output file (default is "goSc4n-summary.txt")
//      --single string           Forced running in single mode
//      --sverbose bool           Store verbose info in summary file
//  -N  --no-output bool          Disable store output
//      --json bool               Store output as JSON format
//      --chunk bool              Enable chunk running against big input
//  -I, --inline string           Inline Detections
//  -q, --quiet                   Enable Quiet Output
//  -Q, --quietFormat string      Format for quiet output (default "{{.VulnURL}}")
//  -R, --report string           HTML report file name
//      --title string            HTML report title
//      --html string             Enable generate HTML reports after the scan done
//      --hh string               Full help message
//      --re                      Shortcut for disable replicate request (avoid sending many request to timeout)
//      --at                      Enable Always True Detection for observe response
//      --lc                      Shortcut for '--proxy http://127.0.0.1:8080'
//      --ba                      Shortcut for take raw input as '{{.BaseURL}}'
//`
	h += "\n\nExamples Commands:\n"
	h += "  goSc4n scan -s <signature> -u <url>\n"
	//h += "  goSc4n scan -c 50 -s <signature> -U <list_urls> -L <level-of-signatures>\n"
	h += "  goSc4n scan -c 50 -s <signature> -U <list_urls>\n"
	//h += "  goSc4n scan -c 50 -s <signature> -U <list_urls> -f 'noti_slack \"{{.vulnInfo}}\"'\n"
	h += "  goSc4n scan -v -c 50 -s <signature> -U list_target.txt -o /tmp/output\n"
	h += "  goSc4n scan -s <signature> -s <another-selector> -u http://example.com\n"
	h += "  goSc4n server -s <signature> -c -v"
	h += "  goSc4n report -o <output directory> --report <Name File>"
	h += "  goSc4n fuzz --site <target> --concurrent <number of threads> --depth 10\n"
	h += "  goSc4n spider --domain <target>\n"
	//h += "  echo '{\"BaseURL\":\"https://example.com/sub/\"}' | goSc4n scan -s sign.yaml -J \n"
	//h += "  goSc4n scan -G -s <signature> -s <another-selector> -x <exclude-selector> -u http://example.com\n"
	//h += "  cat list_target.txt | goSc4n scan -c 100 -s <signature>\n"

	//h += "\nOthers Commands:\n"
	//h += "  goSc4n server -s '/tmp/custom-signature/sensitive/.*' -L 2\n"
	//h += "  goSc4n server --host 0.0.0.0 --port 5000 -s '/tmp/custom-signature/sensitive/.*' -L 2\n"
	//h += "  goSc4n config reload --signDir /tmp/standard-signatures/\n"
	//h += "  goSc4n config add -B /tmp/custom-active-signatures/\n"
	//h += "  goSc4n config update --repo https://github.com/goSc4n-project/goSc4n-signatures\n"
	//h += "  goSc4n report -o /tmp/scanned/out\n"
	//h += "  goSc4n report -o /tmp/scanned/out --title 'Verbose Report' --sverbose\n"
	fmt.Println(h)
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))

}

// HelpMessage print help message
func HelpMessage() {
	h := `
Usage:
  goSc4n config [action]

Config Command examples:
  # Init default signatures
  goSc4n config init

  # Add custom signatures from folder
  goSc4n config add --signDir ~/custom-signatures/

  # Clean old stuff
  goSc4n config clean

  # More examples
  goSc4n config add --signDir /tmp/standard-signatures/
  goSc4n config cred --user sample --pass not123456
	`
	fmt.Println(h)
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))

}

func ScanHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	fmt.Println(cmd.UsageString())
	ScanMessage()
}

func FuzzHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	fmt.Println(cmd.UsageString())
	FuzzMessage()
}

func SpiderHelp(cmd *cobra.Command, _ []string)  {
	fmt.Println(libs.Banner())
	fmt.Println(cmd.UsageString())
	SpiderMessage()
}

func SpiderMessage()  {
	h := "\noptional arguments:\n"
	h += "  --help            show this help message and exit\n"
	h += "  --domain DOMAIN\n"
	h += "                        Domain name of the taget [ex : hackerone.com]\n"
	h += "  --subs SUBS  Set False for no subs [ex : --subs False ]\n"
	h += "  --level LEVEL\n"
	h += "                        For nested parameters [ex : --level high]\n"
	h += "  --exclude EXCLUDE\n"
	h += "                        extensions to exclude [ex --exclude php,aspx]\n"
	h += "  --output OUTPUT\n"
	h += "                        Output file name [by default it is 'domain.txt']\n"
	h += "  --placeholder PLACEHOLDER\n"
	h += "                        The string to add as a placeholder after the parameter\n"
	h += "                        name.\n"
	h += "  --quiet           Do not print the results to the screen\n"
	h += "  --retries RETRIES\n"
	h += "                        Specify number of retries for 4xx and 5xx errors\n"
	h += "\n\nspider --domain http://testphp.vulnweb.com/\n"
	fmt.Println(h)
}

func FuzzMessage()  {
	h := "Flags:\n"
	h += "\t--site string            Site to crawl\n"
	h += "\t--sites string           Site list to crawl\n"
	h += "\t--output string          Output folder\n"
	h += "\t--threads int            Number of threads (Run sites in parallel) (default 1)\n"
	h += "\t--concurrent int         The number of the maximum allowed concurrent requests of the matching domains (default 5)\n"
	h += "\t--depth int              MaxDepth limits the recursion depth of visited URLs. (Set it to 0 for infinite recursion) (default 1)\n"
	h += "\t--quiet                  Suppress all the output and only show URL\n\n"
	h += "\tExample: fuzz --site \"http://testphp.vulnweb.com/\" --concurrent 10 --depth 10\n"
	fmt.Println(h)
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// ScanMessage print help message
func ScanMessage() {
	h := "\nScan Usage example:\n"
	h += "  goSc4n scan -s <signature> -u <url>\n"
	//h += "  goSc4n scan -c 50 -s <signature> -U <list_urls> -L <level-of-signatures>\n"
	h += "  goSc4n scan -c 50 -s <signature> -U <list_urls>\n"
	//h += "  goSc4n scan -c 50 -s <signature> -U <list_urls> -f 'noti_slack \"{{.vulnInfo}}\"'\n"
	h += "  goSc4n scan -v -c 50 -s <signature> -U list_target.txt -o /tmp/output\n"
	h += "  goSc4n scan -s <signature> -s <another-selector> -u http://example.com\n"
	//h += "  echo '{\"BaseURL\":\"https://example.com/sub/\"}' | goSc4n scan -s sign.yaml -J \n"
	//h += "  goSc4n scan -G -s <signature> -s <another-selector> -x <exclude-selector> -u http://example.com\n"
	//h += "  cat list_target.txt | goSc4n scan -c 100 -s <signature>\n"

	h += "\n\nExamples:\n"
	h += "  goSc4n scan -s 'jira' -s 'ruby' -u target.com\n"
	h += "  goSc4n scan -c 50 -s 'java' -x 'tomcat' -U list_of_urls.txt\n"
	h += "  goSc4n scan -G -c 50 -s '/tmp/custom-signature/.*' -U list_of_urls.txt\n"
	h += "  goSc4n scan -v -s '~/my-signatures/products/wordpress/.*' -u 'https://wp.example.com' -p 'root=[[.URL]]'\n"
	h += "  cat urls.txt | grep 'interesting' | goSc4n scan -L 5 -c 50 -s 'fuzz/.*' -U list_of_urls.txt --proxy http://127.0.0.1:8080\n"
	h += "\n"
	fmt.Println(h)
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// ServerHelp report help message
func ServerHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	fmt.Println(cmd.UsageString())
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))

}

// ReportHelp report help message
func ReportHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(libs.Banner())
	fmt.Println(cmd.UsageString())
	fmt.Printf("Official Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

func getEnv(options *libs.Options) {
	if utils.GetOSEnv("GOSC4N_REPO") != "GOSC4N_REPO" {
		options.Config.Repo = utils.GetOSEnv("GOSC4N_REPO")
	}
	if utils.GetOSEnv("GOSC4N_KEY") != "GOSC4N_KEY" {
		options.Config.PrivateKey = utils.GetOSEnv("GOSC4N_KEY")
	}
}

// CleanOutput clean the output folder in case nothing found
func CleanOutput() {
	// clean output
	if utils.DirLength(options.Output) == 0 {
		os.RemoveAll(options.Output)
	}

	// unique vulnSummary
	// Sort sort content of a file
	data := utils.ReadingFileUnique(options.SummaryVuln)
	if len(data) == 0 {
		return
	}
	sort.Strings(data)
	content := strings.Join(data, "\n")
	// remove blank line
	content = regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(content), "\n")
	utils.WriteToFile(options.SummaryVuln, content)
}
