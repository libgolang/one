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

// Flags represents an flags object
type Flags struct {
	subs  map[string]*flag.FlagSet
	props *properties.Properties
}

// NewFlags constructor
func NewFlags() *Flags {

	//
	// Resolve Configuration File
	//
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(&voidWriter{})
	configFilePtr := fs.String("config", os.Getenv("CONFIG"), "Path to configuration file")
	_ = fs.Parse(os.Args)
	if *configFilePtr == "" {
		*configFilePtr = "config.properties"
		if _, err := os.Stat(*configFilePtr); os.IsNotExist(err) {
			cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			*configFilePtr = path.Join(cwd, *configFilePtr)
		}
	}
	props, _ = properties.LoadFile(*configFilePtr, properties.UTF8)

	//
	//
	f := &Flags{
		make(map[string]*flag.FlagSet),
		props,
	}

	return f
}

func (f *Flags) String(key, def, help string) *string {
	return flag.String(key, EnvGet(key, f.props.GetString(key, def)), help)
}

// SubString sub command string
func (f *Flags) SubString(sub, key, def, help string) *string {
	fs, ok := f.subs[sub]
	if !ok {
		fs = flag.NewFlagSet(sub, flag.ContinueOnError)
		f.subs[sub] = fs
	}
	return fs.String(key, EnvGet(key, f.props.GetString(key, def)), help)
}

// Parse parse the commands
func (f *Flags) Parse() {
	if len(os.Args) < 2 {
		f.PrintHelp()
		os.Exit(1)
	}
	if len(f.subs) > 0 {
		fmt.Println("here 3")
		subCommand := os.Args[1]
		if fs, ok := f.subs[subCommand]; ok {
			fmt.Println("here 3.2")
			//
			args := []string{}
			if len(os.Args) > 2 {
				fmt.Println("here 3.3")
				args = os.Args[2:]
			}
			//
			if err := fs.Parse(args); err != nil {
				fmt.Println("here 5")
				f.PrintHelp()
				os.Exit(1)
			}
		} else {
			fmt.Println("here 6")
			f.PrintHelp()
			os.Exit(1)
		}
	} else {
		fmt.Println("here 7")
		f.PrintHelp()
		os.Exit(1)
	}
}

// PrintHelp prints all the flags help
func (f *Flags) PrintHelp() {
	flag.PrintDefaults()
	for _, fl := range f.subs {
		fl.PrintDefaults()
	}
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
