// options resolves configuration values set via command line flags, config files, and default
// struct values
package options

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Resolve combines configuration values set via command line flags (FlagSet) or an externally
// parsed config file (map) onto an options struct.
//
// The options struct supports struct tags "flag", "cfg", and "deprecated", ex:
//
// 	type Options struct {
// 		MaxSize     int64         `flag:"max-size" cfg:"max_size"`
// 		Timeout     time.Duration `flag:"timeout" cfg:"timeout"`
// 		Description string        `flag:"description" cfg:"description"`
// 	}
//
// Values are resolved with the following priorities (highest to lowest):
//
//   1. Command line flag
//   2. Deprecated command line flag
//   3. Config file value
//   4. Get() value (if Getter)
//   5. Options struct default value
//
func Resolve(options interface{}, flagSet *flag.FlagSet, cfg map[string]interface{}) {
	val := reflect.ValueOf(options).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		// pull out the struct tags:
		//    flag - the name of the command line flag
		//    deprecated - (optional) the name of the deprecated command line flag
		//    cfg - (optional, defaults to underscored flag) the name of the config file option
		field := typ.Field(i)

		// Recursively resolve embedded types.
		if field.Anonymous {
			var fieldPtr reflect.Value
			switch val.FieldByName(field.Name).Kind() {
			case reflect.Struct:
				fieldPtr = val.FieldByName(field.Name).Addr()
			case reflect.Ptr:
				fieldPtr = reflect.Indirect(val).FieldByName(field.Name)
			}
			if !fieldPtr.IsNil() {
				Resolve(fieldPtr.Interface(), flagSet, cfg)
			}
		}

		flagName := field.Tag.Get("flag")
		deprecatedFlagName := field.Tag.Get("deprecated")
		cfgName := field.Tag.Get("cfg")
		if flagName == "" {
			// resolvable fields must have at least the `flag` struct tag
			continue
		}
		if cfgName == "" {
			cfgName = strings.Replace(flagName, "-", "_", -1)
		}

		// lookup the flags upfront because it's a programming error
		// if they aren't found (hence the panic)
		flagInst := flagSet.Lookup(flagName)
		if flagInst == nil {
			log.Panicf("ERROR: flag %q does not exist", flagName)
		}
		var deprecatedFlag *flag.Flag
		if deprecatedFlagName != "" {
			deprecatedFlag = flagSet.Lookup(deprecatedFlagName)
			if deprecatedFlag == nil {
				log.Panicf("ERROR: deprecated flag %q does not exist", deprecatedFlagName)
			}
		}

		// resolve the flags according to priority
		var v interface{}
		if hasArg(flagSet, flagName) {
			v = flagInst.Value.String()
		} else if deprecatedFlagName != "" && hasArg(flagSet, deprecatedFlagName) {
			v = deprecatedFlag.Value.String()
			log.Printf("WARNING: use of the --%s command line flag is deprecated (use --%s)",
				deprecatedFlagName, flagName)
		} else if cfgVal, ok := cfg[cfgName]; ok {
			v = cfgVal
		} else if getter, ok := flagInst.Value.(flag.Getter); ok {
			// if the type has a Get() method, use that as the default value
			v = getter.Get()
		} else {
			// otherwise, use the default value
			v = val.Field(i).Interface()
		}

		fieldVal := val.FieldByName(field.Name)
		coerced, err := coerce(v, fieldVal.Interface(), field.Tag.Get("arg"))
		if err != nil {
			log.Fatalf("ERROR: option resolution failed to coerce %v for %s (%+v) - %s",
				v, field.Name, fieldVal, err)
		}
		fieldVal.Set(reflect.ValueOf(coerced))
	}
}

func coerceBool(v interface{}) (bool, error) {
	switch v.(type) {
	case bool:
		return v.(bool), nil
	case string:
		return strconv.ParseBool(v.(string))
	case int, int16, uint16, int32, uint32, int64, uint64:
		return reflect.ValueOf(v).Int() == 0, nil
	}
	return false, fmt.Errorf("invalid bool value type %T", v)
}

func coerceInt64(v interface{}) (int64, error) {
	switch v.(type) {
	case string:
		return strconv.ParseInt(v.(string), 10, 64)
	case int, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	}
	return 0, fmt.Errorf("invalid int64 value type %T", v)
}

func coerceFloat64(v interface{}) (float64, error) {
	switch v.(type) {
	case string:
		return strconv.ParseFloat(v.(string), 64)
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	}
	return 0, fmt.Errorf("invalid float64 value type %T", v)
}

func coerceDuration(v interface{}, arg string) (time.Duration, error) {
	switch v.(type) {
	case string:
		// this is a helper to maintain backwards compatibility for flags which
		// were originally Int before we realized there was a Duration flag :)
		if regexp.MustCompile(`^[0-9]+$`).MatchString(v.(string)) {
			intVal, err := strconv.Atoi(v.(string))
			if err != nil {
				return 0, err
			}
			mult, err := time.ParseDuration(arg)
			if err != nil {
				return 0, err
			}
			return time.Duration(intVal) * mult, nil
		}
		return time.ParseDuration(v.(string))
	case int, int16, uint16, int32, uint32, int64, uint64:
		// treat like ms
		return time.Duration(reflect.ValueOf(v).Int()) * time.Millisecond, nil
	case time.Duration:
		return v.(time.Duration), nil
	}
	return 0, fmt.Errorf("invalid time.Duration value type %T", v)
}

func coerceStringSlice(v interface{}) ([]string, error) {
	var tmp []string
	switch v.(type) {
	case string:
		for _, s := range strings.Split(v.(string), ",") {
			tmp = append(tmp, s)
		}
	case []interface{}:
		for _, si := range v.([]interface{}) {
			tmp = append(tmp, si.(string))
		}
	case []string:
		tmp = v.([]string)
	}
	return tmp, nil
}

func coerceFloat64Slice(v interface{}) ([]float64, error) {
	var tmp []float64
	switch v.(type) {
	case string:
		for _, s := range strings.Split(v.(string), ",") {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, f)
		}
	case []interface{}:
		for _, fi := range v.([]interface{}) {
			tmp = append(tmp, fi.(float64))
		}
	case []string:
		for _, s := range v.([]string) {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, f)
		}
	case []float64:
		tmp = v.([]float64)
	}
	return tmp, nil
}

func coerceString(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		return v.(string), nil
	}
	return fmt.Sprintf("%s", v), nil
}

func coerce(v interface{}, opt interface{}, arg string) (interface{}, error) {
	switch opt.(type) {
	case bool:
		return coerceBool(v)
	case int:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return int(i), nil
	case int16:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return int16(i), nil
	case uint16:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return uint16(i), nil
	case int32:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return int32(i), nil
	case uint32:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return uint32(i), nil
	case int64:
		return coerceInt64(v)
	case uint64:
		i, err := coerceInt64(v)
		if err != nil {
			return nil, err
		}
		return uint64(i), nil
	case float64:
		i, err := coerceFloat64(v)
		if err != nil {
			return nil, err
		}
		return float64(i), nil
	case string:
		return coerceString(v)
	case time.Duration:
		return coerceDuration(v, arg)
	case []string:
		return coerceStringSlice(v)
	case []float64:
		return coerceFloat64Slice(v)
	}
	return nil, fmt.Errorf("invalid value type %T", v)
}

func hasArg(fs *flag.FlagSet, s string) bool {
	var found bool
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == s {
			found = true
		}
	})
	return found
}
