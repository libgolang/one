package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"github.com/libgolang/one/service"
	"github.com/libgolang/one/utils"

	"github.com/libgolang/log"
)

var (
	cfgMasterCertFile    = utils.ConfigString("tls.cert.file", "", "Certificate to use for master REST TLS")
	cfgMasterKeyFile     = utils.ConfigString("tls.key.file", "", "Key file to use for master REST TLS")
	nodeName             = utils.ConfigString("node.name", utils.ResolveNodeName(), "Machine Node Name. Defaults to hostname")
	preRunHookPtr        = utils.ConfigString("hook.run.pre", "", "Pre run hook")
	postRunHookPtr       = utils.ConfigString("hook.run.post", "", "Post run hook")
	proxyBaseDomain      = utils.ConfigStringRequired("proxy.domain", "Proxy Domain")
	proxyPublicIP        = utils.ConfigString("proxy.ip.public", "127.0.0.1", "Public IP address exposed to the internet.")
	proxyPrivateIP       = utils.ConfigString("proxy.ip.private", "127.0.0.1", "Proxy IP address exposed internally only.")   // "10.10.10.1"
	dockerHostIP         = utils.ConfigString("docker.host.ip", "127.0.0.1", "IP address to attach docker port mappings to.") //"10.10.10.1"
	dockerAPIHost        = utils.ConfigString("docker.host", "", "Host or unix that node service uses to connect to docker daemon. E.g.: \"\" = unix socket || \"tcp:/127.0.0.1:2375\"")
	dockerAPIVersion     = utils.ConfigString("docker.api.version", "1.37", "Docker API Version. defaults to 1.37.")
	defDir               = utils.ConfigString("var.dir", "./var", "Var directory.")
	cfgMasterAddrPtr     = utils.ConfigString("master", "", "Starts the master and attaches it to the given address. e.g. --master=127.0.0.1:8080")
	cfgNodeMasterAddrPtr = utils.ConfigString("node", "", "Starts the node and takes the master address. e.g. --node=127.0.0.1:8080")
	db                   service.Db
	dbBack               service.Db
	proxy                service.Proxy
	docker               service.Docker
	cycle                service.Cycle
)

func main() {
	utils.ConfigParse()

	dbBack = service.NewDb(*defDir)
	db = service.NewFrontDb(dbBack)
	proxy = service.NewProxy(*proxyPublicIP, *proxyPrivateIP, *proxyBaseDomain)
	docker = service.NewDocker(*dockerHostIP /*, dockerAPIHost, dockerAPIVersion*/, db)
	cycle = service.NewLifecycle(*dockerHostIP, db, proxy, docker)

	// Init
	rand.Seed(time.Now().Unix())
	defer db.Close()

	// logging
	_ = os.Setenv("LOG_CONFIG", "config.properties")
	log.LoadLogProperties()
	log.SetTrace(true)

	//
	if *cfgMasterAddrPtr == "" && *cfgNodeMasterAddrPtr == "" {
		utils.ConfigPrintHelp()
		os.Exit(1)
	}

	var rs service.RestServer
	if *cfgMasterAddrPtr != "" {
		rs = service.NewRestServer(*cfgMasterAddrPtr, *cfgMasterCertFile, *cfgMasterKeyFile)
		service.NewMasterService(rs, db)
		rs.Start()
	}

	if *cfgNodeMasterAddrPtr != "" {
		service.NewNodeService(*cfgNodeMasterAddrPtr, docker, *nodeName, *dockerHostIP, *preRunHookPtr, *postRunHookPtr)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for sig := range c {
		log.Warn("Captured signal: %d", sig)
		if rs != nil {
			rs.Stop()
		}
		os.Exit(1)
	}

	/*
		case "service":
		case "list":
			listAction(os.Args)
		case "start":
			startService(os.Args)
		case "stop":
			stopService(os.Args)
		default:
			printHelp()
			os.Exit(1)
		}
	*/
}

func listAction(args []string) {
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
	{{.progName}} service --node=127.0.0.1:8080 --master=127.0.0.1:8080
`
