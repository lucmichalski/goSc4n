package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)


func init()  {
	var spiderCmd = &cobra.Command{
		Use: "spider",
		Short: "link tool paramSpider to fuzz website",
		RunE: runParamSpider,
	}

	spiderCmd.Flags().StringVar(&SpiderOp.Domain, "domain", "", "Domain name of the taget [ex : hackerone.com]")
	spiderCmd.Flags().StringVar(&SpiderOp.Level, "level", "", "For nested parameters [ex : --level high]")
	spiderCmd.Flags().StringVar(&SpiderOp.Exclude, "exclude", "", "extensions to exclude [ex --exclude php,aspx]")
	spiderCmd.Flags().StringVar(&SpiderOp.Output, "output", "", "Output file name [by default it is 'domain.txt]'")
	spiderCmd.Flags().StringVar(&SpiderOp.Placeholder, "placeholder", "", "The string to add as a placeholder after the parameter name.")
	spiderCmd.Flags().BoolVar(&SpiderOp.Quiet,"quiet",false,"Do not print the results to the screen")
	spiderCmd.SetHelpFunc(SpiderHelp)
	RootCmd.AddCommand(spiderCmd)
}

func genParamCmd(cmd *cobra.Command){
	if SpiderOp.Domain, _ = cmd.Flags().GetString("domain"); SpiderOp.Domain!=""{
		cmdInput = cmdInput + " -d " +SpiderOp.Domain+ ""
	}

	if SpiderOp.Quiet,_= cmd.Flags().GetBool("quiet"); SpiderOp.Quiet {
		cmdInput = cmdInput + " -q "
	}

	if SpiderOp.Level, _ = cmd.Flags().GetString("level"); SpiderOp.Level != "" {
		cmdInput = cmdInput + " -l " + SpiderOp.Level + " "
	}

	if SpiderOp.Exclude, _ = cmd.Flags().GetString("exclude"); SpiderOp.Exclude != ""{
		cmdInput = cmdInput + " -e " + SpiderOp.Exclude + " "
	}

	if SpiderOp.Output,_ = cmd.Flags().GetString("output"); SpiderOp.Output != ""{
		cmdInput = cmdInput + " -o " + SpiderOp.Output + " "
	}

	if SpiderOp.Placeholder,_ = cmd.Flags().GetString("placeholder"); SpiderOp.Placeholder != ""{
		cmdInput = cmdInput + " -o " + SpiderOp.Placeholder + " "
	}

}

func runParamSpider(cmd *cobra.Command, _ []string)  error{
	cmdInput = ""
	genParamCmd(cmd)
	spiderCmd := exec.Command("bash", "-c", "./crawler/ParamSpider/paramspider.py "+cmdInput+" --output input/test.txt")
	out, err := spiderCmd.Output()
	if err != nil {
		fmt.Println("StdoutPipe: " + err.Error())
	}
	fmt.Println(string(out))
	return nil
}