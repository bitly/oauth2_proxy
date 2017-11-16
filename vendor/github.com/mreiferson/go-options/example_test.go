package options_test

import (
	"flag"
	"fmt"
	"time"

	"github.com/mreiferson/go-options"
)

type Options struct {
	MaxSize     int64         `flag:"max-size" cfg:"max_size"`
	Timeout     time.Duration `flag:"timeout" cfg:"timeout"`
	Description string        `flag:"description" cfg:"description"`
}

func ExampleResolve() {
	flagSet := flag.NewFlagSet("example", flag.ExitOnError)
	flagSet.Int64("max-size", 1024768, "maximum size")
	flagSet.Duration("timeout", 1*time.Hour, "timeout setting")
	flagSet.String("description", "", "description info")
	// parse command line arguments here
	// flagSet.Parse(os.Args[1:])
	flagSet.Parse([]string{"-timeout=5s"})

	opts := &Options{
		MaxSize: 1,
		Timeout: time.Second,
	}
	cfg := map[string]interface{}{
		"timeout": "1h",
	}

	fmt.Printf("%#v", opts)
	options.Resolve(opts, flagSet, cfg)
	fmt.Printf("%#v", opts)
}
