package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rhamerica/one/service"
	"github.com/rhamerica/one/utils"

	"github.com/rhamerica/go-log"

	"github.com/magiconair/properties"
)

var (
	props            = properties.MustLoadFile("config.properties", properties.UTF8)
	proxyBaseDomain  = props.MustGetString("proxy.domain")
	proxyPublicIP    = props.GetString("proxy.ip.public", "127.0.0.1")
	proxyPrivateIP   = props.GetString("proxy.ip.private", "127.0.0.1") // "10.10.10.1"
	dockerHostIP     = props.GetString("docker.host.ip", "127.0.0.1")   //"10.10.10.1"
	dockerAPIHost    = props.GetString("docker.host", "")               // "" = unix socket || "tcp:/127.0.0.1:2375"
	dockerAPIVersion = props.GetString("docker.api.version", "1.37")
	defDir           = props.GetString("var.dir", "./var")
	db               = service.NewDb(defDir)
	proxy            = service.NewProxy(proxyPublicIP, proxyPrivateIP, proxyBaseDomain)
	docker           = service.NewDocker(dockerHostIP, dockerAPIHost, dockerAPIVersion, db)
	cycle            = service.NewLifecycle(dockerHostIP, db, proxy, docker)
)

func main() {
	log.SetTrace(true)
	log.SetDefaultLevel(log.DEBUG)

	n := len(os.Args)
	if n < 2 {
		printHelp()
		os.Exit(1)
	}

	action := os.Args[1]
	switch action {
	case "list":
		listAction(os.Args)
	case "start":
		startService(os.Args)
	case "stop":
		stopService(os.Args)
	default:
		printHelp()
	}
}

func listAction(args []string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", "=Name=", "=State="))
	for _, def := range db.ListDefinitions() {
		isRunning := docker.IsRunningByDefName(def.Name)
		var runningStr string
		if isRunning {
			runningStr = "running"
		} else {
			runningStr = "stopped"
		}
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", def.Name, runningStr))
	}
	_ = w.Flush()
}

func startService(args []string) {
	if len(args) < 3 {
		printHelp()
		return
	}
	serviceName := args[2]

	def, err := db.GetDefinition(serviceName)
	if err != nil {
		fmt.Printf("Service definition %s not found", serviceName)
	}
	cycle.Start(def)
}

func stopService(args []string) {
	if len(args) < 3 {
		printHelp()
		return
	}
	serviceName := args[2]

	def, err := db.GetDefinition(serviceName)
	if err != nil {
		fmt.Printf("Service definition %s not found", serviceName)
	}
	cycle.Stop(def)
}

func printHelp() {
	progName := os.Args[0]
	help := utils.NewTemplate(helpTxt).Context().Set("progName", progName).ParseToString()
	fmt.Println(help)
}

const helpTxt = `
	{{.progName}} list
	{{.progName}} start <name>
	{{.progName}} stop  <name>
`
