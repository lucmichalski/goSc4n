package core

import (
	"encoding/base64"
	"fmt"
	"github.com/goSc4n/goSc4n/libs"
	"github.com/goSc4n/goSc4n/utils"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"math/rand"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
	"github.com/dop251/goja"
)

// ParseVariable parse variable in YAML signature file
func ParseVariable(sign libs.Signature) []map[string]string {
	var realVariables []map[string]string
	rawVariables := make(map[string][]string)
	// reading variable
	for _, variable := range sign.Variables {
		for key, value := range variable {
			// strip out blank line
			if strings.Trim(value, " ") == "" {
				continue
			}

			// variable as a script
			if strings.Contains(value, "(") && strings.Contains(value, ")") {
				if strings.Contains(value, "{{.") && strings.Contains(value, "}}") {
					value = ResolveVariable(value, sign.Target)
				}
				rawVariables[key] = RunVariables(value)
			}
			/*
				- variable: [google.com,example.com]
			*/
			// variable as a list
			if strings.HasPrefix(value, "[") && strings.Contains(value, ",") {
				rawVar := strings.Trim(value[1:len(value)-1], " ")
				rawVariables[key] = strings.Split(rawVar, ",")
				continue
			}
			/*
				- variable: |
					google.com
					example.com
			*/
			if strings.Contains(value, "\n") {
				value = strings.Trim(value, "\n\n")
				rawVariables[key] = strings.Split(value, "\n")
				continue
			}
		}
	}

	if len(rawVariables) == 1 {
		for k, v := range rawVariables {
			for _, value := range v {
				variable := make(map[string]string)
				variable[k] = value
				realVariables = append(realVariables, variable)
			}
		}
		return realVariables
	}

	// select max number of list
	var maxLength int
	for _, v := range rawVariables {
		if maxLength < len(v) {
			maxLength = len(v)
		}
	}

	// @TODO: Need to improve this
	if len(rawVariables) > 1 && len(rawVariables) <= 3 {
		keys := funk.Keys(rawVariables).([]string)
		list1 := rawVariables[keys[0]]
		list2 := rawVariables[keys[1]]

		if len(rawVariables) == 2 {
			for _, item1 := range list1 {
				// loop in second var
				for _, item2 := range list2 {
					element := make(map[string]string)
					element[keys[0]] = item1
					element[keys[1]] = item2
					realVariables = append(realVariables, element)
				}
			}
		} else if len(rawVariables) == 3 {
			list3 := rawVariables[keys[2]]
			for _, item1 := range list1 {
				// loop in second var
				for _, item2 := range list2 {
					// loop in third var
					for _, item3 := range list3 {
						element := make(map[string]string)
						element[keys[0]] = item1
						element[keys[1]] = item2
						element[keys[2]] = item3
						realVariables = append(realVariables, element)
					}
				}
			}
		}

		//fmt.Println("realVariables: ", realVariables)
		//fmt.Println("len(realVariables): ", len(realVariables))
		return realVariables
	}

	// make all variable to same length
	Variables := make(map[string][]string)
	for k, v := range rawVariables {
		Variables[k] = utils.ExpandLength(v, maxLength)
	}

	// join all together to make list of map variable
	for i := 0; i < maxLength; i++ {
		for j := 0; j < maxLength; j++ {
			variable := make(map[string]string)
			for k, v := range Variables {
				variable[k] = v[j]
			}
			realVariables = append(realVariables, variable)
		}
	}

	seen := make(map[string]bool)
	// just unique the variables
	var uniqVariables []map[string]string
	for index := 0; index < len(realVariables); index++ {
		for k, v := range realVariables[index] {
			val := fmt.Sprintf("%v%v", k, v)
			if _, ok := seen[val]; !ok {
				// fmt.Println(k, v)
				seen[val] = true
				uniqVariables = append(uniqVariables, realVariables[index])
			}
		}
	}

	return uniqVariables
}

// RunVariables is main function for detections
func RunVariables(variableString string) []string {
	var extra []string
	if !strings.Contains(variableString, "(") {
		return extra
	}

	vm := goja.New()


	vm.Set("Content", func(call goja.FunctionCall) goja.Value {
		filename := call.Argument(0).String()
		filename = utils.NormalizePath(filename)
		data := utils.GetFileContent(filename)
		if len(data) > 0 {
			extra = append(extra, data)
		}
		return vm.ToValue("")
	})

	vm.Set("InputCmd", func(call goja.FunctionCall) goja.Value {
		cmd := call.Argument(0).String()
		data := InputCmd(cmd)
		if len(data) <= 0 {

		}else if !strings.Contains(data, "\n") {
			extra = append(extra, data)
		}else{
			extra = append(extra, strings.Split(data, "\n")...)
		}
		return vm.ToValue("")
	})

	vm.Set("Range", func(call goja.FunctionCall) goja.Value {
		min, err := strconv.Atoi(call.Argument(0).String())
		max, err := strconv.Atoi(call.Argument(1).String())
		if err == nil {
			for i := min; i < max; i++ {
				extra = append(extra, fmt.Sprintf("%v", i))
			}
		}
		return vm.ToValue("")
	})

	vm.Set("SplitLines", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).String()
		extra = append(extra, SplitLines(data)...)
		return vm.ToValue("")
	})

	vm.Set("Base64Encode", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).String()
		extra = append(extra, Base64Encode(data))
		return vm.ToValue("")
	})

	vm.Set("Base64Decode", func(call goja.FunctionCall) goja.Value {
		raw := call.Argument(0).String()
		data, err := base64.StdEncoding.DecodeString(raw)
		if err == nil {
			extra = append(extra, string(data))
		}
		return vm.ToValue("")
	})


	vm.Set("URLEncode", func(call goja.FunctionCall) goja.Value {
		data := call.Argument(0).String()
		extra = append(extra, URLEncode(data))
		return vm.ToValue("")
	})

	vm.Set("URLEncodeByLines", func(call goja.FunctionCall) goja.Value {
		data := SplitLines(call.Argument(0).String())
		if len(data) != 0 {
			for _, line := range data {
				extra = append(extra, URLEncode(line))
			}
		}
		return vm.ToValue("")
	})

	utils.DebugF("variableString: %v", variableString)
	vm.RunString(variableString)
	return extra
}

// RandomString return a random string with length
func RandomString(n int) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[seededRand.Intn(len(letter))]
	}
	return string(b)
}

// InputCmd take input as os command
// @NOTE: this is a feature not an RCE :P
func InputCmd(Cmd string) string {
	command := []string{
		"bash",
		"-c",
		Cmd,
	}
	out, _ := exec.Command(command[0], command[1:]...).CombinedOutput()
	return strings.TrimSpace(string(out))
}

// SplitLines just split new Line
func SplitLines(raw string) []string {
	var result []string
	if strings.Contains(raw, "\n") {
		result = strings.Split(raw, "\n")
	}
	return result
}

// Base64Encode just Base64 Encode
func Base64Encode(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// URLEncode just URL Encode
func URLEncode(raw string) string {
	return url.QueryEscape(raw)
}

// GenPorts gen list of ports based on input
func GenPorts(raw string) []string {
	var ports []string
	if strings.Contains(raw, ",") {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if strings.Contains(item, "-") {
				min, err := strconv.Atoi(strings.Split(item, "-")[0])
				if err != nil {
					continue
				}
				max, err := strconv.Atoi(strings.Split(item, "-")[1])
				if err != nil {
					continue
				}
				for i := min; i <= max; i++ {
					ports = append(ports, fmt.Sprintf("%v", i))
				}
			} else {
				ports = append(ports, item)
			}
		}
	} else {
		if strings.Contains(raw, "-") {
			min, err := strconv.Atoi(strings.Split(raw, "-")[0])
			if err != nil {
				return ports
			}
			max, err := strconv.Atoi(strings.Split(raw, "-")[1])
			if err != nil {
				return ports
			}
			for i := min; i <= max; i++ {
				ports = append(ports, fmt.Sprintf("%v", i))
			}
		} else {
			ports = append(ports, raw)
		}
	}

	return ports
}

// ReplicationJob replication more jobs based on the signature
func ReplicationJob(input string, sign libs.Signature) ([]libs.Job, error) {
	var jobs []libs.Job

	u, err := url.Parse(input)
	// something wrong so parsing it again
	if err != nil || u.Scheme == "" || strings.Contains(u.Scheme, ".") {
		input = fmt.Sprintf("https://%v", input)
		u, err = url.Parse(input)
		if err != nil {
			return jobs, fmt.Errorf("error parsing input: %s", input)
		}
	}

	var urls, ports, prefiixes []string
	if sign.Replicate.Ports != "" {
		ports = GenPorts(sign.Replicate.Ports)
	}

	if strings.TrimSpace(sign.Replicate.Prefixes) != "" {
		value := sign.Replicate.Prefixes
		// variable as a script
		if strings.Contains(value, "(") && strings.Contains(value, ")") {
			if strings.Contains(value, "{{.") && strings.Contains(value, "}}") {
				value = ResolveVariable(value, sign.Target)
			}
			prefiixes = append(prefiixes, RunVariables(value)...)
		}
		/*
			- variable: foo,bar
		*/
		// variable as a list
		if strings.Contains(value, ",") {
			prefiixes = append(prefiixes, strings.Split(strings.TrimSpace(value), ",")...)
		}
		/*
			- variable: |
				google.com
				example.com
		*/
		if strings.Contains(value, "\n") {
			value = strings.Trim(value, "\n\n")
			prefiixes = append(prefiixes, strings.Split(value, "\n")...)

		}
	}

	if len(ports) > 0 {
		for _, port := range ports {
			cloneURL := url.URL{}
			err = copier.Copy(&cloneURL, u)
			if err != nil {
				continue
			}
			oPort := cloneURL.Port()
			nPort := fmt.Sprintf(":%s", port)
			if oPort == "" {
				cloneURL.Host += nPort
			} else {
				// avoid duplicate port here
				if strings.Contains(cloneURL.Host, nPort) {
					continue
				}
				cloneURL.Host = strings.Replace(cloneURL.Host, fmt.Sprintf(":%s", oPort), nPort, -1)
			}

			urlWithPort := cloneURL.String()
			urls = append(urls, urlWithPort)
		}
	}

	if len(prefiixes) > 0 {
		if len(urls) == 0 {
			urls = append(urls, input)
		}

		for _, urlRaw := range urls {
			u, err := url.Parse(urlRaw)
			if err != nil {
				continue
			}

			for _, prefix := range prefiixes {
				prefix = strings.TrimSpace(prefix)
				cloneURL := url.URL{}
				err = copier.Copy(&cloneURL, u)
				if err != nil {
					continue
				}

				cloneURL.Path = path.Join(cloneURL.Path, prefix)
				urlWithPrefix := cloneURL.String()
				if !sign.BasePath {
					urlWithPrefix = fmt.Sprintf("%s://%s/%s", cloneURL.Scheme, cloneURL.Host, prefix)
				}
				urls = append(urls, urlWithPrefix)
			}
		}
	}

	for _, urlRaw := range urls {
		job := libs.Job{
			URL:  urlRaw,
			Sign: sign,
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
