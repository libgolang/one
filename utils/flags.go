package utils

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
)

var (
	emptyString = ""
	flagMap     = make(map[string]*flagDef)
	loadedProps *properties.Properties
)

type flagDef struct {
	name     string
	defValue string
	usage    string
	value    string
	valuePtr *string
	required bool
}

// ConfigString defines a string flag
func ConfigString(name, defValue, usage string) *string {
	props := getConfig()
	def := &flagDef{
		name:     name,
		defValue: props.GetString(name, defValue),
		usage:    usage,
		required: false,
	}
	flagMap[name] = def
	return &def.value
}

// ConfigStringRequired defines a required string flag
func ConfigStringRequired(name, usage string) *string {
	props := getConfig()
	def := &flagDef{
		name:     name,
		defValue: props.GetString(name, ""),
		usage:    usage,
		required: true,
	}
	flagMap[name] = def
	return &def.value
}

// ConfigParse call parse on flags
func ConfigParse() {
	for _, v := range flagMap {
		v.valuePtr = flag.String(v.name, v.defValue, v.usage)
	}
	flag.Parse()
	for _, v := range flagMap {
		if *v.valuePtr == "" && v.required {
			panic(fmt.Sprintf("\n%s is a required field.\n\n", v.name))
		}
		v.value = *v.valuePtr
	}
}

// ConfigPrintHelp prints flag helps
func ConfigPrintHelp() {
	flag.PrintDefaults()
}

// KeyToEnvKey renames a configuration key to a proper environment variable
func KeyToEnvKey(key string) string {
	envKey := strings.Replace(key, ".", "_", -1)
	envKey = strings.Replace(key, "-", "_", -1)
	envKey = strings.ToUpper(key)
	return envKey
}

// EnvGet gets environment variable
func EnvGet(key, def string) string {
	v, ok := os.LookupEnv(KeyToEnvKey(key))
	if !ok {
		v = def
	}
	return v
}

//
type voidWriter struct {
}

func (w *voidWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func getConfig() *properties.Properties {

	if loadedProps != nil {
		return loadedProps
	}

	// get it from environment
	envFileName, ok := os.LookupEnv("CONFIG")
	if !ok {
		envFileName = "config.properties"
		if _, err := os.Stat(envFileName); os.IsNotExist(err) {
			cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			envFileName = path.Join(cwd, envFileName)
		}
	}

	// get it from flags
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(&voidWriter{})
	configFilePtr := fs.String("config", envFileName, "Path to configuration file")
	_ = fs.Parse(os.Args[1:])
	loadedProps, _ = properties.LoadFile(*configFilePtr, properties.UTF8)
	return loadedProps
}
