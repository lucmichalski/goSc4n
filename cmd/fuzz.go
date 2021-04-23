package cmd

import (
	"fmt"
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

	fuzzCmd.Flags().StringVar(&options.Fuzz.Site, "site", "", "Site to crawl")
	fuzzCmd.Flags().StringVar(&options.Fuzz.Sites,"sites", "", "Site list to crawl")
	fuzzCmd.Flags().StringVar(&options.Fuzz.Output,"output", "", "Output folder")
	fuzzCmd.Flags().IntVar(&options.Fuzz.Threads,"threads", 1,"Number of threads (Run sites in parallel)")
	fuzzCmd.Flags().IntVar(&options.Fuzz.Concurrent,"concurrent",  5, "The number of the maximum allowed concurrent requests of the matching domains")
	fuzzCmd.Flags().IntVar(&options.Fuzz.Depth,"depth", 1, "MaxDepth limits the recursion depth of visited URLs. (Set it to 0 for infinite recursion)")
	fuzzCmd.Flags().BoolVar(&options.Fuzz.Quiet,"quiet", true, "Suppress all the output and only show URL")
	RootCmd.AddCommand(fuzzCmd)
}

func genCmd(cmd *cobra.Command) {
	if options.Fuzz.Quiet,_= cmd.Flags().GetBool("quiet"); options.Fuzz.Quiet {
		cmdInput = cmdInput + " -q "
	}

	if options.Fuzz.Concurrent,_=cmd.Flags().GetInt("concurrent");options.Fuzz.Concurrent != 5 && options.Fuzz.Concurrent > 0{
		cmdInput= cmdInput + " -c " + strconv.Itoa(options.Fuzz.Concurrent)
	}

	if options.Fuzz.Depth,_=cmd.Flags().GetInt("depth");options.Fuzz.Depth != 1 && options.Fuzz.Depth > 0{
		cmdInput= cmdInput + " -d " + strconv.Itoa(options.Fuzz.Depth)
	}

	if options.Fuzz.Threads,_=cmd.Flags().GetInt("threads");options.Fuzz.Threads != 1 && options.Fuzz.Threads > 0{
		cmdInput= cmdInput + " -t " + strconv.Itoa(options.Fuzz.Threads)
	}

	if options.Fuzz.Output,_=cmd.Flags().GetString("output");options.Fuzz.Output != ""{
		cmdInput= cmdInput + " -o "+options.Fuzz.Output+" "
	}

	if options.Fuzz.Sites,_=cmd.Flags().GetString("sites");options.Fuzz.Sites != ""{
		cmdInput=cmdInput + " -S \""+options.Fuzz.Sites+"\" "
	}
	if options.Fuzz.Site,_=cmd.Flags().GetString("site");options.Fuzz.Site != ""{
		cmdInput=cmdInput + " -s \""+options.Fuzz.Site+"\" "
	}
}

func runSpider(cmd *cobra.Command, _ []string) error {
	helps, _ := cmd.Flags().GetBool("hh")
	if helps == true {
		FuzzHelp()
		os.Exit(1)
	}
	cmdInput = ""
	genCmd(cmd)
	spiderCmd := exec.Command("bash", "-c", "./crawler/gospider "+cmdInput +" > input/fuzzOutput.txt")
	out, err := spiderCmd.Output()
	if err != nil {
		fmt.Println("StdoutPipe: " + err.Error())
	}
	fmt.Println(string(out))
	fmt.Println("You can get output in: ./input/fuzzOutput")
	return nil
}