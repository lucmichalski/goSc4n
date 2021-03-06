package core

import (
	"fmt"
	"github.com/goSc4n/goSc4n/libs"
	"github.com/goSc4n/goSc4n/sender"
	"github.com/goSc4n/goSc4n/utils"
	"strings"
)

// Runner runner struct
type Runner struct {
	Input       string
	SendingType string
	Opt         libs.Options
	Sign        libs.Signature
	Origin      Record
	Target      map[string]string
	Records     []Record
}

// Record all information about request
type Record struct {
	// main part
	Request  libs.Request
	Response libs.Response
	Sign     libs.Signature

	OriginReq libs.Request
	OriginRes libs.Response
	Origins   []libs.Origin
	// for output
	Opt         libs.Options
	RawOutput   string
	ExtraOutput string
	// for detection
	PassCondition bool
	IsVulnerable  bool
	DetectString  string
	DetectResult  string
	ScanID        string
}

//
//func InitRunnerWithDefaultOpt(url string, sign string) {
//}

// InitRunner init task
func InitRunner(url string, sign libs.Signature, opt libs.Options) (Runner, error) {
	var runner Runner
	runner.Input = url
	runner.Opt = opt
	runner.Sign = sign
	runner.SendingType = "parallels"
	runner.PrepareTarget()

	if runner.Sign.Single || runner.Sign.Serial {
		runner.SendingType = "serial"
	}

	// sending origin if we have it here
	if runner.Sign.Origin.Method != "" || runner.Sign.Origin.Res != "" {
		runner.PrePareOrigin()
	}

	// generate requests
	runner.GetRequests()
	return runner, nil
}

func (r *Runner) PrepareTarget() {
	// clean up the '//' on hostname in case we use --ba option
	if r.Opt.Mics.BaseRoot || r.Sign.CleanSlash {
		r.Input = strings.TrimRight(r.Input, "/")
	}

	Target := make(map[string]string)
	// parse Input from JSON format
	if r.Opt.EnableFormatInput {
		Target = ParseInputFormat(r.Input)
	} else {
		Target = ParseTarget(r.Input)
	}

	if r.Opt.Mics.BaseRoot || r.Sign.Replicate.Prefixes != "" {
		Target["BaseURL"] = Target["Raw"]
	}

	r.Sign.Target = Target
	r.Target = Target
}

// GetRequests get requests ready to send
func (r *Runner) GetRequests() {
	reqs := r.GenRequests()
	if len(reqs) > 0 {
		for _, req := range reqs {
			var rec Record
			// set somethings in record
			rec.Request = req
			rec.Request.Target = r.Target
			rec.Sign = r.Sign
			rec.Opt = r.Opt
			// assign origins here
			rec.OriginReq = r.Origin.Request
			rec.OriginRes = r.Origin.Response

			r.Records = append(r.Records, rec)
		}
	}
}

// GenRequests generate request for sending
func (r *Runner) GenRequests() []libs.Request {
	// quick param for calling resource
	r.Sign.Target = MoreVariables(r.Sign.Target, r.Sign, r.Opt)

	var realReqs []libs.Request
	globalVariables := ParseVariable(r.Sign)
	if len(globalVariables) > 0 {
		for _, globalVariable := range globalVariables {
			r.Sign.Target = r.Target
			for k, v := range globalVariable {
				r.Sign.Target[k] = v
			}
			// start to send stuff
			for _, req := range r.Sign.Requests {
				// receive request from "-r req.txt"
				if r.Sign.RawRequest != "" {
					req.Raw = r.Sign.RawRequest
				}
				// gen bunch of request to send
				realReqs = append(realReqs, ParseRequest(req, r.Sign, r.Opt)...)
			}
		}
	} else {
		r.Sign.Target = r.Target
		// start to send stuff
		for _, req := range r.Sign.Requests {
			// receive request from "-r req.txt"
			if r.Sign.RawRequest != "" {
				req.Raw = r.Sign.RawRequest
			}
			// gen bunch of request to send
			realReqs = append(realReqs, ParseRequest(req, r.Sign, r.Opt)...)
		}
	}
	return realReqs
}

func (r *Runner) PrePareOrigin() {
	var originRec libs.Record
	var origin libs.Origin
	// prepare initial signature and variables
	Target := make(map[string]string)
	Target = MoreVariables(r.Target, r.Sign, r.Opt)
	// base origin
	if r.Sign.Origin.Method != "" || r.Sign.Origin.Res != "" {
		origin, Target = r.SendOrigin(r.Sign.Origin)
		originRec.Request = origin.ORequest
		originRec.Response = origin.OResponse
	}

	// in case we have many origin
	if len(r.Sign.Origins) > 0 {
		var origins []libs.Origin
		for index, origin := range r.Sign.Origins {
			origin, Target = r.SendOrigin(origin.ORequest)
			if origin.Label == "" {
				origin.Label = fmt.Sprintf("%v", index)
			}
			origins = append(origins, origin)
		}
		r.Sign.Origins = origins
	}

	r.Target = Target
}

// sending origin request
func (r *Runner) SendOrigin(originReq libs.Request) (libs.Origin, map[string]string) {
	var origin libs.Origin
	var err error
	var originRes libs.Response

	originSign := r.Sign
	if r.Opt.Scan.RawRequest != "" {
		RawRequest := utils.GetFileContent(r.Opt.Scan.RawRequest)
		originReq = ParseBurpRequest(RawRequest)
	}

	if originReq.Raw == "" {
		originSign.Target = r.Target
		originReq = ParseOrigin(originReq, originSign, r.Opt)
	}

	// parse response directly without sending
	if originReq.Res != "" {
		originRes = ParseBurpResponse("", originReq.Res)
	} else {
		originRes, err = sender.JustSend(r.Opt, originReq)
		if err == nil {
			if r.Opt.Verbose && (originReq.Method != "") {
				fmt.Printf("[Sent-Origin] %v %v %v %v %v\n", originReq.Method, originReq.URL, originRes.Status, originRes.ResponseTime, len(originRes.Beautify))
			}
		}
	}

	originRec := Record{Request: originReq, Response: originRes}
	// set some more variables
	originRec.Conclude()

	for k, v := range originSign.Target {
		if r.Target[k] == "" {
			r.Target[k] = v
		}
	}

	origin.ORequest = originReq
	origin.OResponse = originRes
	r.Origin = originRec
	return origin, r.Target
}
