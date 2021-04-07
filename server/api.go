package server

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/goSc4n/goSc4n/tree/hoangnm/core"
	"github.com/goSc4n/goSc4n/tree/hoangnm/database"
	"github.com/goSc4n/goSc4n/tree/hoangnm/database/models"
	"github.com/goSc4n/goSc4n/tree/hoangnm/libs"
	"github.com/goSc4n/goSc4n/tree/hoangnm/utils"
	"github.com/panjf2000/ants"
	"github.com/thoas/go-funk"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Ping testing authenticated connection
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "200",
		"message": "pong",
	})
}

// GetStats return stat data
func GetStats(c *gin.Context) {
	var info []models.Record
	database.DB.Where("risk = ?", "Info").Find(&info)
	var potential []models.Record
	database.DB.Where("risk = ?", "Potential").Find(&potential)
	var low []models.Record
	database.DB.Where("risk = ?", "Low").Find(&low)
	var medium []models.Record
	database.DB.Where("risk = ?", "Medium").Find(&medium)
	var high []models.Record
	database.DB.Where("risk = ?", "High").Find(&high)
	var critical []models.Record
	database.DB.Where("risk = ?", "Critical").Find(&critical)

	stats := []int{
		len(info),
		len(potential),
		len(low),
		len(medium),
		len(high),
		len(critical),
	}

	c.JSON(200, gin.H{
		"status":  "200",
		"message": "Success",
		"stats":   stats,
	})
}

// GetSignSummary return signature stat
func GetSignSummary(c *gin.Context) {
	var signs []models.Signature
	var categories []string
	var data []int
	database.DB.Find(&signs).Pluck("DISTINCT category", &categories)
	// stats := make(map[string]int)
	for _, category := range categories {
		var signatures []models.Signature
		database.DB.Where("category = ?", category).Find(&signatures)
		data = append(data, len(signatures))
	}

	c.JSON(200, gin.H{
		"status":     "200",
		"message":    "Success",
		"categories": categories,
		"data":       data,
	})
}

// GetSigns return signature record
func GetSigns(c *gin.Context) {
	var signs []models.Signature
	database.DB.Find(&signs)

	c.JSON(200, gin.H{
		"status":     "200",
		"message":    "Success",
		"signatures": signs,
	})
}

// GetAllScan return all scans
func GetAllScan(c *gin.Context) {
	var scans []models.Scans
	database.DB.Find(&scans)

	// remove empty scan
	var realScans []models.Scans
	for _, scan := range scans {
		var rec models.Record
		database.DB.First(&rec, "scan_id = ?", scan.ScanID)
		//if rec.ScanID != "" {
			realScans = append(realScans, scan)
		//}
	}

	c.JSON(200, gin.H{
		"status":  "200",
		"message": "Success",
		"scans":   realScans,
	})
}

// GetRecords get record by scan ID
func GetRecords(c *gin.Context) {
	sid := c.Param("sid")
	var records []models.Record
	database.DB.Where("scan_id = ?", sid).Find(&records)

	c.JSON(200, gin.H{
		"status":  "200",
		"message": "Success",
		"records": records,
	})
}

// GetRecord get record detail by record ID
func GetRecord(c *gin.Context) {
	rid := c.Param("rid")
	var record models.Record
	database.DB.Where("id = ?", rid).First(&record)

	c.JSON(200, gin.H{
		"status":  "200",
		"message": "Success",
		"record":  record,
	})
}

// SignConfig config
type SignConfig struct {
	Value string `json:"sign"`
}

// UpdateDefaultSign geet record by scan
func UpdateDefaultSign(c *gin.Context) {
	var signConfig SignConfig
	err := c.ShouldBindJSON(&signConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	database.UpdateDefaultSign(signConfig.Value)
	c.JSON(200, gin.H{
		"status":  "200",
		"message": "Update Defeult sign success",
	})
}

type ScanDTO struct{
	Url	string	`json:"host" binding:"required"`
	Signatures []string `json:"signatures"`
	Num	int	`json:"num"`
}

type DataDTO struct{
	ScanDto ScanDTO	`json:"data"`
}


func addScan(options libs.Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataDTO DataDTO
		err := c.BindJSON(&dataDTO)
		if err != nil {
			log.Fatalln(err)
			c.Status(http.StatusBadRequest)
		}else{
			SelectSign(options)
			var urls []string
			if dataDTO.ScanDto.Url != "" {
				urls = append(urls, dataDTO.ScanDto.Url)
			}
			options.Signs = dataDTO.ScanDto.Signatures
			var wg sync.WaitGroup
			p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
				CreateRunner(i,options)
				wg.Done()
			}, ants.WithPreAlloc(true))
			defer p.Release()

			for _, signFile := range options.SelectedSigns {
				sign, err := core.ParseSign(signFile)
				if err != nil {
					utils.ErrorF("Error parsing YAML sign: %v", signFile)
					continue
				}
				// filter signature by level
				if sign.Level > options.Level {
					continue
				}

				// Submit tasks one by one.
				for _, url := range urls {
					wg.Add(1)
					job := libs.Job{URL: url, Sign: sign}
					_ = p.Invoke(job)
				}
			}

			wg.Wait()
			CleanOutput(options)

			c.JSON(200,gin.H{
				"status":"200",
				"mess":"Success",
			})
		}
	}
}

func CleanOutput(options libs.Options) {
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

func CreateRunner(j interface{},options libs.Options) {
	job := j.(libs.Job)
	jobs := []libs.Job{job}

	if (job.Sign.Replicate.Ports != "" || job.Sign.Replicate.Prefixes != "") && !options.Mics.DisableReplicate {
		if options.Mics.BaseRoot {
			job.Sign.BasePath = true
		}
		moreJobs, err := core.ReplicationJob(job.URL, job.Sign)
		if err == nil {
			jobs = append(jobs, moreJobs...)
		}
	}

	for _, job := range jobs {
		runner, err := core.InitRunner(job.URL, job.Sign, options)
		if err != nil {
			utils.ErrorF("Error create new runner: %v", err)
		}
		runner.Sending()
	}
}

func SelectSign(options libs.Options) {
	var selectedSigns []string
	// read selector from File
	if options.Selectors != "" {
		options.Signs = append(options.Signs, utils.ReadingFileUnique(options.Selectors)...)
	}

	// default is all signature
	if len(options.Signs) == 0 {
		selectedSigns = core.SelectSign("**")
	}

	// search signature through Signatures table
	for _, signName := range options.Signs {
		selectedSigns = append(selectedSigns, core.SelectSign(signName)...)
		if !options.NoDB {
			Signs := database.SelectSign(signName)
			selectedSigns = append(selectedSigns, Signs...)
		}
	}
	options.SelectedSigns = selectedSigns

	if len(selectedSigns) == 0 {
		fmt.Fprintf(os.Stderr, "[Error] No signature loaded\n")
		fmt.Fprintf(os.Stderr, "Try '%s' to init default signatures\n", color.GreenString("goSc4n config init"))
		os.Exit(1)
	}
	selectedSigns = funk.UniqString(selectedSigns)
	utils.InforF("Signatures Loaded: %v", len(selectedSigns))
	signInfo := fmt.Sprintf("Signature Loaded: ")
	for _, signName := range selectedSigns {
		signInfo += fmt.Sprintf("%v ", filepath.Base(signName))
	}
	utils.InforF(signInfo)

	// create new scan or group with old one
	var scanID string
	if options.ScanID == "" {
		scanID = database.NewScan(options, "scan", selectedSigns)
	} else {
		scanID = options.ScanID
	}
	utils.InforF("Start Scan with ID: %v", scanID)
	options.ScanID = scanID
}