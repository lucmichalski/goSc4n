package core

import (
	"github.com/goSc4n/goSc4n/tree/hoangnm/utils"
	//"github.com/robertkrimen/otto"
	"os/exec"
	"github.com/dop251/goja"
)

// Conclude is main function for detections
func (r *Record) MiddleWare() {
	//record := *r
	vm := goja.New()
	var middlewareOutput string

	vm.Set("InvokeCmd", func(call goja.FunctionCall) goja.Value {
		rawCmd := call.Argument(0).String()
		result := InvokeCmd(r, rawCmd)
		middlewareOutput += result
		utils.DebugF(result)
		return vm.ToValue("")
	})


	for _, middleScript := range r.Request.Middlewares {
		utils.DebugF("[MiddleWare]: %s", middleScript)
		vm.RunString(middleScript)
	}
	r.Request.MiddlewareOutput = middlewareOutput
}

// InvokeCmd execute external command
func InvokeCmd(rec *Record, rawCmd string) string {
	target := ParseTarget(rec.Request.URL)
	realCommand := Encoder(rec.Request.Encoding, ResolveVariable(rawCmd, target))
	utils.DebugF("Execute Command: %v", realCommand)
	command := []string{
		"bash",
		"-c",
		realCommand,
	}
	out, _ := exec.Command(command[0], command[1:]...).CombinedOutput()
	rec.Request.MiddlewareOutput = string(out)
	return string(out)
}