package main

import (
	//"flag"
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
	//cfgMasterAddr     = utils.ConfigString("master", "")
	cfgMasterCertFile = utils.ConfigString("tls.cert.file", "")
	cfgMasterKeyFile  = utils.ConfigString("tls.key.file", "")
	//cfgNodeMasterAddr = utils.ConfigString("node", "")

	nodeName         = utils.ConfigString("node.name", "")
	proxyBaseDomain  = utils.ConfigRequireString("proxy.domain")
	proxyPublicIP    = utils.ConfigString("proxy.ip.public", "127.0.0.1")
	proxyPrivateIP   = utils.ConfigString("proxy.ip.private", "127.0.0.1") // "10.10.10.1"
	dockerHostIP     = utils.ConfigString("docker.host.ip", "127.0.0.1")   //"10.10.10.1"
	dockerAPIHost    = utils.ConfigString("docker.host", "")               // "" = unix socket || "tcp:/127.0.0.1:2375"
	dockerAPIVersion = utils.ConfigString("docker.api.version", "1.37")
	defDir           = utils.ConfigString("var.dir", "./var")

	dbBack = service.NewDb(defDir)
	db     = service.NewFrontDb(dbBack)
	proxy  = service.NewProxy(proxyPublicIP, proxyPrivateIP, proxyBaseDomain)
	docker = service.NewDocker(dockerHostIP /*, dockerAPIHost, dockerAPIVersion*/, db)
	cycle  = service.NewLifecycle(dockerHostIP, db, proxy, docker)
)

func main() {

	// flag
	f := utils.NewFlags()
	cfgMasterAddrPtr := f.SubString("service", "master", "", "Starts the master and attaches it to the given address. e.g. --master=127.0.0.1:8080")
	cfgNodeMasterAddrPtr := f.SubString("service", "node", "", "Starts the node and takes the master address. e.g. --node=127.0.0.1:8080")
	f.Parse()

	// Init
	rand.Seed(time.Now().Unix())
	defer db.Close()

	// logging
	os.Setenv("LOG_CONFIG", "config.properties")
	log.LoadLogProperties()
	log.SetTrace(true)
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")

	//log.SetDefaultLevel(log.DEBUG)
	//log.SetWriters(append([]log.Writer{}, log.NewFileWriter(path.Join(defDir, "logs"), "one", log.Gigabyte, 10)))

	//serviceCommand := flag.NewFlagSet("service", flag.ExitOnError)
	//masterTextPtr := serviceCommand.String("master", cfgMasterAddr, "Starts the master and attaches it to the given address. e.g. --master=127.0.0.1:8080")
	//nodeTextPtr := serviceCommand.String("node", cfgNodeMasterAddr, "Starts the node and takes the master address. e.g. --node=127.0.0.1:8080")

	//listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	//startCommand := flag.NewFlagSet("start", flag.ExitOnError)
	//stopCommand := flag.NewFlagSet("stop", flag.ExitOnError)

	// args
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "service":
		//_ = serviceCommand.Parse(os.Args[2:])
		cfgMasterAddr := *cfgMasterAddrPtr
		cfgNodeMasterAddr := *cfgNodeMasterAddrPtr
		if cfgMasterAddr == "" && cfgNodeMasterAddr == "" {
			fmt.Println("Here B")
			f.PrintHelp()
			os.Exit(1)
		}

		var rs service.RestServer
		if cfgMasterAddr != "" {
			fmt.Println("Here C")
			rs = service.NewRestServer(cfgMasterAddr, cfgMasterCertFile, cfgMasterKeyFile)
			service.NewMasterService(rs, db)
			rs.Start()
		}

		if cfgNodeMasterAddr != "" {
			fmt.Println("Here D")
			service.NewNodeService(cfgNodeMasterAddr, docker, utils.ResolveNodeName(nodeName), dockerHostIP)
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

	case "list":
		//_ = listCommand.Parse(os.Args[2:])
		listAction(os.Args)
	case "start":
		//_ = startCommand.Parse(os.Args[2:])
		startService(os.Args)
	case "stop":
		//_ = stopCommand.Parse(os.Args[2:])
		stopService(os.Args)
	default:
		printHelp()
		os.Exit(1)
	}

	/*
		if serviceCommand.Parsed() {
		} else if listCommand.Parsed() {
		} else if startCommand.Parsed() {
		} else if stopCommand.Parsed() {
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
