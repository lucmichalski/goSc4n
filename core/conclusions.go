package core

import (


	"github.com/dop251/goja"
	"github.com/goSc4n/goSc4n/utils"

	//"github.com/robertkrimen/otto"
	"os/exec"
	"regexp"
	"strings"
)

// check conditions before sending request
func (r *Record) Condition() bool {
	check := r.RequestScripts("condition", r.Request.Conditions)
	return check
}

// Conclude is main function for detections
func (r *Record) Conclude() {
	record := *r
	vm := goja.New()

	// ExecCmd execute command command
	vm.Set("ExecCmd", func(call goja.FunctionCall) goja.Value {
		result := vm.ToValue(Execution(call.Argument(0).String()))
		return result
	})

	// write something to a file
	vm.Set("WriteTo", func(call goja.FunctionCall) goja.Value {
		dest := utils.NormalizePath(call.Argument(0).String())
		value := call.Argument(1).String()
		utils.WriteToFile(dest, value)
		return vm.ToValue("")
	})

	vm.Set("StringSearch", func(call goja.FunctionCall) goja.Value {
		componentName := call.Argument(0).String()
		analyzeString := call.Argument(1).String()
		component := GetComponent(record, componentName)
		validate := StringSearch(component, analyzeString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("StringCount", func(call goja.FunctionCall) goja.Value {
		componentName := call.Argument(0).String()
		analyzeString := call.Argument(1).String()
		component := GetComponent(record, componentName)
		validate := StringCount(component, analyzeString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("RegexSearch", func(call goja.FunctionCall) goja.Value {
		componentName := call.Argument(0).String()
		analyzeString := call.Argument(1).String()
		component := GetComponent(record, componentName)
		_, validate := RegexSearch(component, analyzeString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("RegexCount", func(call goja.FunctionCall) goja.Value {
		componentName := call.Argument(0).String()
		analyzeString := call.Argument(1).String()
		component := GetComponent(record, componentName)
		validate := RegexCount(component, analyzeString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("StatusCode", func(call goja.FunctionCall) goja.Value {
		statusCode := record.Response.StatusCode
		result:= vm.ToValue(statusCode)
		return result
	})
	vm.Set("ResponseTime", func(call goja.FunctionCall) goja.Value {
		responseTime := record.Response.ResponseTime
		result := vm.ToValue(responseTime)
		return result
	})
	vm.Set("ContentLength", func(call goja.FunctionCall) goja.Value {
		ContentLength := record.Response.Length
		result := vm.ToValue(ContentLength)
		return result
	})

	// StringSelect select a string from component
	// e.g: StringSelect("component", "res1", "right", "left")
	vm.Set("StringSelect", func(call goja.FunctionCall) goja.Value {
		componentName := call.Argument(0).String()
		valueName := call.Argument(1).String()
		left := call.Argument(2).String()
		right := call.Argument(3).String()
		component := GetComponent(record, componentName)
		value := Between(component, left, right)
		r.Request.Target[valueName] = value
		utils.DebugF("StringSelect: %v --> %v", valueName, value)
		return vm.ToValue("")
	})

	//  - RegexSelect("component", "regex")
	//  - RegexSelect("component", "regex")
	vm.Set("RegexSelect", func(call goja.FunctionCall) goja.Value {
		result := RegexSelect(record, call.Arguments)
		if len(result) > 0 {
			for k, value := range result {
				utils.DebugF("New variales: %v -- %v", k, value)
				r.Request.Target[k] = value
			}
		}
		return vm.ToValue("")
	})

	// SetValue("var_name", StatusCode())
	// SetValue("status", StringCount('middleware', '11'))
	vm.Set("SetValue", func(call goja.FunctionCall) goja.Value {
		valueName := call.Argument(0).String()
		value := call.Argument(1).String()
		utils.DebugF("SetValue: %v -- %v", valueName, value)
		r.Request.Target[valueName] = value
		return vm.ToValue("")
	})

	for _, concludeScript := range record.Request.Conclusions {
		utils.DebugF("[Conclude]: %v", concludeScript)
		vm.RunString(concludeScript)
	}
}

// Between get string between left and right
func Between(value string, left string, right string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, left)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, right)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(left)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

// RegexSelect get regex string from component
func RegexSelect(realRec Record, arguments []goja.Value) map[string]string {
	result := make(map[string]string)
	//  - RegexSelect("component", "var_name", "regex")
	utils.DebugF("arguments -- %v", arguments)
	if len(arguments) < 2 {
		utils.DebugF("Invalid Conclude")
		return result
	}
	componentName := arguments[0].String()
	component := GetComponent(realRec, componentName)
	regexString := arguments[1].String()

	// map all selected
	var myExp, err = regexp.Compile(regexString)
	if err != nil {
		utils.ErrorF("Regex error: %v %v", regexString, err)
		return result
	}
	match := myExp.FindStringSubmatch(component)
	if len(match) == 0 {
		utils.DebugF("No match found: %v", regexString)
	}
	for i, name := range myExp.SubexpNames() {
		if i != 0 && name != "" && len(match) > i {
			result[name] = match[i]
		}
	}
	return result
}

// Execution Run a command
func Execution(cmd string) string {
	command := []string{
		"bash",
		"-c",
		cmd,
	}
	var output string
	utils.DebugF("[Exec] %v", command)
	realCmd := exec.Command(command[0], command[1:]...)
	out, _ := realCmd.CombinedOutput()
	output = string(out)
	return output
}
