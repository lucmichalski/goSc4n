package core

import (
	"encoding/hex"
	"fmt"
	"github.com/goSc4n/goSc4n/sender"
	"github.com/goSc4n/goSc4n/utils"

	//"github.com/robertkrimen/otto"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

func (r *Record) Detector() {
	if r.Opt.InlineDetection != "" {
		r.Request.Detections = append(r.Request.Detections, r.Opt.InlineDetection)
	}
	r.RequestScripts("detections", r.Request.Detections)
}

// Detector is main function for detections
func (r *Record) RequestScripts(scriptType string, scripts []string) bool {
	/* Analyze part */
	if r.Request.Beautify == "" {
		r.Request.Beautify = sender.BeautifyRequest(r.Request)
	}
	if len(r.Request.Detections) <= 0 {
		return false
	}

	record := *r
	var extra string
	vm := goja.New()

	// ExecCmd execute command command
	vm.Set("ExecCmd", func(call goja.FunctionCall) goja.Value {
		result := vm.ToValue(Execution(call.Argument(0).String()))
		return result
	})



	vm.Set("StringGrepCmd", func(call goja.FunctionCall) goja.Value {
		command := call.Argument(0).String()
		searchString := call.Argument(0).String()
		result := vm.ToValue(StringSearch(Execution(command), searchString))
		return result
	})

	vm.Set("RegexGrepCmd", func(call goja.FunctionCall) goja.Value {
		command := call.Argument(0).String()
		searchString := call.Argument(0).String()
		_, validate := RegexSearch(Execution(command), searchString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("StringSearch", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		componentName := "response"
		analyzeString := args[0].String()
		if len(args) >= 2 {
			componentName = args[0].String()
			analyzeString = args[1].String()
		}
		component := GetComponent(record, componentName)
		validate := StringSearch(component, analyzeString)
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("search", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		componentName := "response"
		analyzeString := args[0].String()
		if len(args) >= 2 {
			componentName = args[0].String()
			analyzeString = args[1].String()
		}
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
		args := call.Arguments
		componentName := "response"
		analyzeString := args[0].String()
		if len(args) >= 2 {
			componentName = args[0].String()
			analyzeString = args[1].String()
		}
		component := GetComponent(record, componentName)
		matches, validate := RegexSearch(component, analyzeString)
		result := vm.ToValue(validate)
		//if err != nil {
		//	utils.ErrorF("Error Regex: %v", analyzeString)
		//	result, _ = vm.ToValue(false)
		//}
		if matches != "" {
			extra = matches
		}
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
		result := vm.ToValue(statusCode)
		return result
	})


	vm.Set("ResponseTime", func(call goja.FunctionCall) goja.Value {
		responseTime := record.Response.ResponseTime
		result := vm.ToValue(responseTime)
		return result
	})

	vm.Set("time", func(call goja.FunctionCall) goja.Value {
		responseTime := record.Response.ResponseTime
		result := vm.ToValue(responseTime)
		return result
	})


	vm.Set("ContentLength", func(call goja.FunctionCall) goja.Value {
		args := call.Arguments
		if len(args) == 0 {
			ContentLength := record.Response.Length
			result := vm.ToValue(ContentLength)
			return result
		}
		componentName := args[0].String()
		componentLength := len(GetComponent(record, componentName))
		result := vm.ToValue(componentLength)
		return result
	})


	vm.Set("HasPopUp", func(call goja.FunctionCall) goja.Value {
		result := vm.ToValue(record.Response.HasPopUp)
		return result
	})


	//  - RegexGrep("component", "regex")
	//  - RegexGrep("component", "regex", "position")
	vm.Set("RegexGrep", func(call goja.FunctionCall) goja.Value {
		value := RegexGrep(record, call.Arguments)
		result := vm.ToValue(value)
		return result
	})

	// check if folder, file exist or not
	vm.Set("Exist", func(call goja.FunctionCall) goja.Value {
		input := utils.NormalizePath(call.Argument(0).String())
		var exist bool
		if utils.FileExists(input) {
			exist = true
		}
		if utils.FolderExists(input) {
			exist = true
		}
		result := vm.ToValue(exist)
		return result
	})

	vm.Set("DirLength", func(call goja.FunctionCall) goja.Value {
		validate := utils.DirLength(call.Argument(0).String())
		result := vm.ToValue(validate)
		return result
	})

	vm.Set("FileLength", func(call goja.FunctionCall) goja.Value {
		validate := utils.FileLength(call.Argument(0).String())
		result := vm.ToValue(validate)
		return result
	})

	/* Really start do detection here */
	switch scriptType {
	case "detect", "detections":
		for _, analyze := range scripts {
			// pass detection here
			result, _ := vm.RunString(analyze)
			analyzeResult := result.Export()
			// in case vm panic
			//if err != nil || analyzeResult == nil {
			//	r.DetectString = analyze
			//	r.IsVulnerable = false
			//	r.DetectResult = ""
			//	r.ExtraOutput = ""
			//	continue
			//}
			r.DetectString = analyze
			r.IsVulnerable = analyzeResult.(bool)
			r.DetectResult = extra
			r.ExtraOutput = extra

			utils.DebugF("[Detection] %v -- %v", analyze, r.IsVulnerable)
			// deal with vulnerable one here
			next := r.Output()
			if next == "stop" {
				return true
			}
		}
		return r.IsVulnerable
	case "condition", "conditions":
		var valid bool
		for _, analyze := range scripts {
			result, _ := vm.RunString(analyze)
			analyzeResult := result.Export()
			// in case vm panic
			//if err != nil || analyzeResult == nil {
			//	r.PassCondition = false
			//	continue
			//}
			r.PassCondition = analyzeResult.(bool)
			utils.DebugF("[Condition] %v -- %v", analyze, r.PassCondition)
			valid = r.PassCondition
		}
		return valid
	}
	return false
}

// StringSearch search string literal in component
func StringSearch(component string, analyzeString string) bool {
	var result bool
	if strings.Contains(component, analyzeString) {
		result = true
	}
	utils.DebugF("analyzeString: %v -- %v", analyzeString, result)
	return result
}

// StringCount count string literal in component
func StringCount(component string, analyzeString string) int {
	return strings.Count(component, analyzeString)
}

// RegexSearch search regex string in component
func RegexSearch(component string, analyzeString string) (string, bool) {
	var result bool
	var extra string
	r, err := regexp.Compile(analyzeString)
	if err != nil {
		return extra, result
	}

	matches := r.FindStringSubmatch(component)
	if len(matches) > 0 {
		result = true
		extra = strings.Join(matches, "\n")
	}
	utils.DebugF("Component: %v", component)
	utils.DebugF("analyzeRegex: %v -- %v", analyzeString, result)
	return extra, result
}

// RegexCount count regex string in component
func RegexCount(component string, analyzeString string) int {
	r, err := regexp.Compile(analyzeString)
	if err != nil {
		return 0
	}
	matches := r.FindAllStringIndex(component, -1)
	return len(matches)
}

// RegexGrep grep regex string from component
func RegexGrep(realRec Record, arguments []goja.Value) string {
	componentName := arguments[0].String()
	component := GetComponent(realRec, componentName)

	regexString := arguments[1].String()
	var position int
	var err error
	if len(arguments) > 2 {
		position, err = strconv.Atoi(arguments[2].String())
		if err != nil {
			position = 0
		}
	}

	var value string
	r, rerr := regexp.Compile(regexString)
	if rerr != nil {
		return ""
	}
	matches := r.FindStringSubmatch(component)
	if len(matches) > 0 {
		if position <= len(matches) {
			value = matches[position]
		} else {
			value = matches[0]
		}
	}
	return value
}

// GetComponent get component to run detection
func GetComponent(record Record, component string) string {
	component = strings.ToLower(component)
	utils.DebugF("Get Component: %v", component)
	switch component {
	case "orequest":

		return record.OriginReq.Beautify
	case "oresheaders", "oheaders", "ohead", "oresheader":
		beautifyHeader := fmt.Sprintf("%v \n", record.OriginRes.Status)
		for _, header := range record.OriginRes.Headers {
			for key, value := range header {
				beautifyHeader += fmt.Sprintf("%v: %v\n", key, value)
			}
		}
		return beautifyHeader
	case "obody", "oresbody":
		return record.OriginRes.Body
	case "oresponse", "ores":
		return record.OriginRes.Beautify
	case "request":
		return record.Request.Beautify
	case "response":
		if record.Response.Beautify == "" {
			return record.Response.Body
		}
		return record.Response.Beautify
	case "resheader", "resheaders", "headers", "header":
		beautifyHeader := fmt.Sprintf("%v \n", record.Response.Status)
		for _, header := range record.Response.Headers {
			for key, value := range header {
				beautifyHeader += fmt.Sprintf("%v: %v\n", key, value)
			}
		}
		return beautifyHeader
	case "body", "resbody":
		return record.Response.Body
	case "bytes", "byte", "hex":
		return hex.EncodeToString([]byte(record.Request.Beautify))
	case "byteBody", "hexBody":
		return hex.EncodeToString([]byte(record.Request.Body))
	case "middleware":
		return record.Request.MiddlewareOutput
	default:
		return record.Response.Beautify
	}
}
