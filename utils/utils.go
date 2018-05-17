package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"flag"

	"github.com/libgolang/log"
	"github.com/magiconair/properties"
)

// FileExists check if file exists
func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// Mkdir create directory
func Mkdir(file string) {
	if err := os.MkdirAll(file, 0775); err != nil {
		panic(err)
	}
}

// Remove file/directory
func Remove(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}

// EnsureDir checks if a directory exists, if it doesn't it attempts to create it
func EnsureDir(dir string) {
	if !FileExists(dir) {
		log.Warn("Directory %s does not exist.  Creating!", dir)
		Mkdir(dir)
	}
}

// ResolveNodeName resolves the name of the current machine.
// It arguments as presedence.  If the arguments are empty,
// then it uses the machine hostname.
func ResolveNodeName(args ...string) string {
	for _, hn := range args {
		if hn != "" {
			return hn
		}
	}
	name, _ := os.Hostname()
	return name
}

// ConfigString Resolves configuration from argument, environment or config file
// in that order
func ConfigString(key, def string) string {
	configInit()
	flagset := flag.NewFlagSet("args", flag.ContinueOnError)
	flagset.SetOutput(&voidWriter{})
	val := flagset.String(key, def, "")
	_ = flagset.Parse(os.Args[1:])
	if *val == "" {
		envKey := strings.Replace(key, ".", "_", -1)
		envKey = strings.Replace(key, "-", "_", -1)
		envKey = strings.ToUpper(key)
		*val = os.Getenv(envKey)
		if *val == "" {
			*val = props.GetString(key, "")
		}
	}
	return *val
}

// ConfigRequireString Resolves configuration from argument, environment or config file
// in that order. If not able to resolve, then help is printed and os.Exit(1) is called.
func ConfigRequireString(key string) string {
	configInit()
	flagset := flag.NewFlagSet("args", flag.ContinueOnError)
	flagset.SetOutput(&voidWriter{})
	val := flagset.String(key, "", "")
	_ = flagset.Parse(os.Args[1:])
	if *val == "" {
		envKey := strings.Replace(key, ".", "_", -1)
		envKey = strings.Replace(key, "-", "_", -1)
		envKey = strings.ToUpper(key)
		*val = os.Getenv(envKey)
		if *val == "" {
			*val = props.GetString(key, "")
		}
	}
	if *val == "" {
		flagset.SetOutput(os.Stderr)
		flagset.PrintDefaults()
		os.Exit(1)
	}
	return *val
}

// used for configuration resolution
var props *properties.Properties

func configInit() {
	if props != nil {
		return
	}
	cfgFlagset := flag.NewFlagSet("config", flag.ContinueOnError)
	cfgFlagset.SetOutput(&voidWriter{})
	configFilePtr := cfgFlagset.String("config", os.Getenv("CONFIG"), "")
	_ = cfgFlagset.Parse(os.Args[1:])
	if *configFilePtr == "" {
		*configFilePtr = "config.properties"
		if _, err := os.Stat(*configFilePtr); os.IsNotExist(err) {
			cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			*configFilePtr = path.Join(cwd, *configFilePtr)
		}
	}
	props, _ = properties.LoadFile(*configFilePtr, properties.UTF8)
}
