package options_test

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/mreiferson/go-options"
)

// TestFlagSetDefaults verifies that default flag values are applied in the
// absence of user-specified setting.
func TestFlagSetDefaults(t *testing.T) {
	flagSet := flag.NewFlagSet("TestFlagSetDefaults", flag.PanicOnError)
	opts := &Options{1024768, 1 * time.Hour, ""}

	flagSet.Int64("max-size", 1024768, "maximum size")
	flagSet.Duration("timeout", 1*time.Hour, "timeout setting")
	flagSet.String("description", "", "description info")

	if err := flagSet.Parse([]string{"-timeout=5s"}); err != nil {
		t.Fatal(err)
	}

	cfg := map[string]interface{}{}

	options.Resolve(opts, flagSet, cfg)

	if expected, actual := flagSet.Lookup("max-size").Value.(flag.Getter).Get().(int64), opts.MaxSize; actual != expected {
		t.Errorf("Expected opts.MaxSize to default to %v but actual=%v", expected, actual)
	}
}

// TestConfigWithOverlappingOsArgs verifies that flag values set via map config
// are respected.
func TestConfigWithOverlappingOsArgs(t *testing.T) {
	osArgsBak := make([]string, len(os.Args))
	for i, arg := range os.Args {
		osArgsBak[i] = arg
	}
	defer func() { os.Args = osArgsBak }() // Restore os.Args afterwards.

	os.Args = []string{"./logserver/logserver", "-config", "/tmp/cfg.toml"}

	type ServerOptions struct {
		Server  bool `flag:"server"`
		Debug   bool `flag:"debug"`
		Version bool `flag:"version"`
	}

	flagSet := flag.NewFlagSet("TestConfigWithOverlappingOsArgs", flag.PanicOnError)

	flagSet.Bool("server", false, "run in server mode")
	flagSet.Bool("debug", false, "toggle debug output")
	flagSet.Bool("version", false, "show version information and then exit")

	if err := flagSet.Parse(os.Args); err != nil {
		t.Fatal(err)
	}

	opts := &ServerOptions{}
	cfg := map[string]interface{}{
		"server": true,
	}

	options.Resolve(opts, flagSet, cfg)

	if expected, actual := true, opts.Server; actual != expected {
		t.Errorf("Expected opts.Server=%v but actual=%v", expected, actual)
	}
	if expected, actual := false, opts.Debug; actual != expected {
		t.Errorf("Expected opts.Debug=%v but actual=%v", expected, actual)
	}
	if expected, actual := false, opts.Version; actual != expected {
		t.Errorf("Expected opts.Version=%v but actual=%v", expected, actual)
	}
}

func TestFloat64(t *testing.T) {
	type ConfigurableThing struct {
		Percentage float64 `flag:"percentage"`
	}

	const defaultValue = 0.5

	testCases := []struct {
		Args     []string
		Expected float64
	}{
		{
			Args:     []string{""},
			Expected: defaultValue,
		},
		{
			Args:     []string{},
			Expected: defaultValue,
		},
		{
			Args:     []string{"-percentage", fmt.Sprint(defaultValue)},
			Expected: defaultValue,
		},
		{
			Args:     []string{"-percentage", "0.753"},
			Expected: 0.753,
		},
		{
			Args:     []string{"-percentage", "-0.753"},
			Expected: -0.753,
		},
		{
			Args:     []string{"-percentage=-0.117983"},
			Expected: -0.117983,
		},
	}

	for i, testCase := range testCases {
		flagSet := flag.NewFlagSet("TestFloat64", flag.PanicOnError)

		flagSet.Float64("percentage", defaultValue, "integer or decimal representing the percentage")

		if err := flagSet.Parse(testCase.Args); err != nil {
			t.Fatal(err)
		}

		configThing := &ConfigurableThing{defaultValue}
		cfg := map[string]interface{}{}

		options.Resolve(configThing, flagSet, cfg)

		if expected, actual := testCase.Expected, configThing.Percentage; actual != expected {
			t.Errorf("[i=%v testCase=%+v] Expected configThing.Percentage=%v but actual=%v", i, testCase, expected, actual)
		}
	}
}

func TestEmbeddedTypeOptions(t *testing.T) {
	newFlagSet := func() *flag.FlagSet {
		flagSet := flag.NewFlagSet("TestEmbeddedTypeOptions", flag.ExitOnError)
		flagSet.Bool("peaceful", false, "peaceful flag details")
		flagSet.Bool("inside", false, "inside flag details")
		flagSet.Bool("undo", false, "undo flag details")
		flagSet.Bool("install", false, "install flag details")
		flagSet.Int("arbitrary", 0, "arbitrary flag details")
		flagSet.String("fields", "", "fields flag details")
		flagSet.String("user", "", "user flag details")
		return flagSet
	}

	type EmbeddedAsValue struct {
		Arbitrary int      `flag:"arbitrary"`
		Fields    []string `flag:"fields"`
	}

	type EmbeddedAsPointer struct {
		EmbeddedAsValue
		Peaceful bool `flag:"peaceful"`
		Inside   bool `flag:"inside"`
		Undo     bool `flag:"undo"`
	}

	type SampleConfig struct {
		*EmbeddedAsPointer
		Install    bool   `flag:"install"`
		User       string `flag:"user"`
		otherField string
	}

	testCases := []struct {
		Args         []string
		CallbackFunc func(config *SampleConfig)
	}{
		{
			Args: []string{"-inside"},
			CallbackFunc: func(config *SampleConfig) {
				if expected, actual := false, config.Peaceful; actual != expected {
					t.Errorf("Expected config.Peaceful=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := true, config.Inside; actual != expected {
					t.Errorf("Expected config.Inside=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := false, config.Undo; actual != expected {
					t.Errorf("Expected config.Undo=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
			},
		},
		{
			Args: []string{"-peaceful", "-undo", "-user=mreiferson"},
			CallbackFunc: func(config *SampleConfig) {
				if expected, actual := true, config.Peaceful; actual != expected {
					t.Errorf("Expected config.Peaceful=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := false, config.Inside; actual != expected {
					t.Errorf("Expected config.Inside=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := true, config.Undo; actual != expected {
					t.Errorf("Expected config.Undo=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := "mreiferson", config.User; actual != expected {
					t.Errorf("Expected config.User=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := 0, config.EmbeddedAsPointer.Arbitrary; actual != expected {
					t.Errorf("Expected config.EmbeddedAsPointer.Arbitrary=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
			},
		},
		{
			Args: []string{"-peaceful", "-arbitrary", "99", "-fields=a,a1,B,B1,c"},
			CallbackFunc: func(config *SampleConfig) {
				if expected, actual := true, config.Peaceful; actual != expected {
					t.Errorf("Expected config.Peaceful=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := 99, config.Arbitrary; actual != expected {
					t.Errorf("Expected config.Arbitrary=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
				if expected, actual := []string{"a", "a1", "B", "B1", "c"}, config.EmbeddedAsPointer.EmbeddedAsValue.Fields; !reflect.DeepEqual(actual, expected) {
					t.Errorf("Expected config.EmbeddedAsPointer.EmbeddedAsValue.Fields=%v but actual=%v (config=%+v)", expected, actual, *config)
				}
			},
		},
	}

	for _, testCase := range testCases {
		var (
			config = &SampleConfig{
				EmbeddedAsPointer: &EmbeddedAsPointer{
					EmbeddedAsValue: EmbeddedAsValue{},
				},
				otherField: "hello there!",
			}
			flagSet = newFlagSet()
		)

		flagSet.Parse(testCase.Args)

		options.Resolve(config, flagSet, nil)

		testCase.CallbackFunc(config)
	}
}
