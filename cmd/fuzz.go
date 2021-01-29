package cmd

import (
	"fmt"
	"github.com/goSc4n/goSc4n/tree/hoangnm/core"
	"github.com/goSc4n/goSc4n/tree/hoangnm/libs"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strconv"
)

var cmdInput string

func init()  {
	var fuzzCmd = &cobra.Command{
		Use: "fuzz",
		Short: "link tool goSpider to fuzz website",
		RunE: runSpider,
	}

	fuzzCmd.Flags().StringP("site", "s", "", "Site to crawl")
	fuzzCmd.Flags().StringP("sites", "S", "", "Site list to crawl")
	fuzzCmd.Flags().StringP("proxy", "p", "", "Proxy (Ex: http://127.0.0.1:8080)")
	fuzzCmd.Flags().StringP("output", "o", "", "Output folder")
	fuzzCmd.Flags().StringP("user-agent", "u", "", "User Agent to use\n\tweb: random web user-agent\n\tmobi: random mobile user-agent\n\tor you can set your special user-agent\")")
	fuzzCmd.Flags().StringArrayP("header", "H", []string{}, "Header to use (Use multiple flag to set multiple header)")

	fuzzCmd.Flags().IntP("threads", "t", 1, "Number of threads (Run sites in parallel)")
	fuzzCmd.Flags().IntP("concurrent", "c", 5, "The number of the maximum allowed concurrent requests of the matching domains")
	fuzzCmd.Flags().IntP("depth", "d", 1, "MaxDepth limits the recursion depth of visited URLs. (Set it to 0 for infinite recursion)")
	fuzzCmd.Flags().IntP("delay", "k", 0, "Delay is the duration to wait before creating a new request to the matching domains (second)")
	fuzzCmd.Flags().IntP("random-delay", "K", 0, "RandomDelay is the extra randomized duration to wait added to Delay before creating a new request (second)")
	fuzzCmd.Flags().IntP("timeout", "m", 10, "Request timeout (second)")

	fuzzCmd.Flags().BoolP("base", "B", false, "Disable all and only use HTML content")
	fuzzCmd.Flags().BoolP("other-source", "a", false, "Find URLs from 3rd party (Archive.org, CommonCrawl.org, VirusTotal.com, AlienVault.com)")
	fuzzCmd.Flags().BoolP("include-subs", "w", false, "Include subdomains crawled from 3rd party. Default is main domain")
	fuzzCmd.Flags().BoolP("include-other-source", "r", false, "Also include other-source's urls (still crawl and request)")

	fuzzCmd.Flags().BoolP("verbose", "v", false, "Turn on verbose")
	fuzzCmd.Flags().BoolP("quiet", "q", false, "Suppress all the output and only show URL")

	fuzzCmd.Flags().SortFlags = false
	if err := fuzzCmd.Execute(); err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	RootCmd.AddCommand(fuzzCmd)
}

func genCmd(cmd *cobra.Command) {
	if &options.Fuzz.Quiet,_= cmd.Flags().GetBool("quiet"); &options.Fuzz.Quiet {
		cmdInput = cmdInput + " -q "
	}

	if &options.Fuzz.Verbose,_=cmd.Flags().GetBool("verbose");&options.Fuzz.Verbose {
		cmdInput = cmdInput + " -v "
	}

	if &options.Fuzz.IncludeOtherSource,_=cmd.Flags().GetBool("include-other-source");&options.Fuzz.IncludeOtherSource {
		cmdInput = cmdInput + " -r "
	}

	if &options.Fuzz.IncludeSubs,_=cmd.Flags().GetBool("include-subs");&options.Fuzz.IncludeSubs {
		cmdInput = cmdInput + " -w "
	}

	if &options.Fuzz.OtherSource,_=cmd.Flags().GetBool("other-source");&options.Fuzz.OtherSource {
		cmdInput = cmdInput + " -a "
	}

	if &options.Fuzz.Base,_=cmd.Flags().GetBool("base");&options.Fuzz.Base {
		cmdInput = cmdInput + " -b "
	}

	if &options.Fuzz.Concurrent,_=cmd.Flags().GetInt("concurrent");&options.Fuzz.Concurrent != 5 && &options.Fuzz.Concurrent > 0{
		cmdInput= cmdInput + " -c " + strconv.Itoa(&options.Fuzz.Concurrent)
	}

	if &options.Fuzz.Depth,_=cmd.Flags().GetInt("depth");&options.Fuzz.Depth != 1 && &options.Fuzz.Depth > 0{
		cmdInput= cmdInput + " -d " + strconv.Itoa(&options.Fuzz.Depth)
	}

	if &options.Fuzz.Threads,_=cmd.Flags().GetInt("threads");&options.Fuzz.Threads != 1 && &options.Fuzz.Threads > 0{
		cmdInput= cmdInput + " -t " + strconv.Itoa(&options.Fuzz.Threads)
	}

	if &options.Fuzz.Delay,_=cmd.Flags().GetInt("delay");&options.Fuzz.Delay != 0 && &options.Fuzz.Delay > 0{
		cmdInput= cmdInput + " -k " + strconv.Itoa(&options.Fuzz.Delay)
	}

	if &options.Fuzz.RandomDelay,_=cmd.Flags().GetInt("random-delay");&options.Fuzz.RandomDelay != 0 && &options.Fuzz.RandomDelay > 0{
		cmdInput= cmdInput + " -K " + strconv.Itoa(&options.Fuzz.RandomDelay)
	}

	if &options.Fuzz.Timeout,_=cmd.Flags().GetInt("time-out");&options.Fuzz.Timeout != 10 && &options.Fuzz.Timeout > 0{
		cmdInput= cmdInput + " -K " + strconv.Itoa(&options.Fuzz.Timeout)
	}

	&options.Fuzz.Header,_=cmd.Flags().GetStringArray("header")
	for i:=0;i < len(&options.Fuzz.Header);i++ {
		cmdInput = cmdInput + " -H \""+ &options.Fuzz.Header[i]+"\" "
	}

	if &options.Fuzz.UserAgent,_=cmd.Flags().GetString("user-agent");&options.Fuzz.UserAgent != "" {
		cmdInput= cmdInput + " -u \"" + &options.Fuzz.UserAgent+"\" "
	}

	if &options.Fuzz.Output,_=cmd.Flags().GetString("output");&options.Fuzz.Output != ""{
		cmdInput= cmdInput + " -o "+&options.Fuzz.Output+" "
	}

	if &options.Fuzz.Proxy,_=cmd.Flags().GetString("proxy");&options.Fuzz.Proxy != ""{
		cmdInput=cmdInput + " -p \""+&options.Fuzz.Proxy+"\" "
	}

	if &options.Fuzz.Sites,_=cmd.Flags().GetString("sites");&options.Fuzz.Sites != ""{
		cmdInput=cmdInput + " -S \""+&options.Fuzz.Sites+"\" "
	}
	if &options.Fuzz.Site,_=cmd.Flags().GetString("site");&options.Fuzz.Site != ""{
		cmdInput=cmdInput + " -s \""+&options.Fuzz.Site+"\" "
	}
}

func runSpider(cmd *cobra.Command, _ []string) error {
	cmdInput = ""
	genCmd(cmd)
	spiderCmd := exec.Command("bash", "-c", "cd /mnt/c/Users/QuangVinh/Desktop/Test && ./gospider "+cmdInput)
	out, err := spiderCmd.Output()
	if err != nil {
		fmt.Println("StdoutPipe: " + err.Error())
	}
	fmt.Println(string(out))
	return nil
}