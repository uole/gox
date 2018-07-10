package option

import (
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func hasArg(fs *flag.FlagSet, s string) bool {
	var found bool
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == s {
			found = true
		}
	})
	return found
}

func Merge(options interface{}, flagSet *flag.FlagSet) error {
	val := reflect.ValueOf(options).Elem()
	typ := val.Type()
	allowTags := []string{"flag", "json", "xml", "toml", "ini"}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			fmt.Println(field.Name)
			var fieldPtr reflect.Value
			switch val.FieldByName(field.Name).Kind() {
			case reflect.Struct:
				fieldPtr = val.FieldByName(field.Name).Addr()
			case reflect.Ptr:
				fieldPtr = reflect.Indirect(val).FieldByName(field.Name)
			}
			if !fieldPtr.IsNil() {
				Merge(fieldPtr.Interface(), flagSet)
			}
		}
		var flagName string
		for _, tag := range allowTags {
			flagName = field.Tag.Get(tag)
			if flagName != "" {
				break
			}
		}
		if flagName == "" {
			continue
		}
		flagName = strings.Replace(flagName, "_", "-", -1)
		flagValue := flagSet.Lookup(flagName)

		//if is default value ,skip merge
		if flagValue == nil || flagValue.DefValue == flagValue.Value.String() {
			continue
		}

		var value interface{}
		if hasArg(flagSet, flagName) {
			value = flagValue.Value.String()
		} else {
			value = val.Field(i).Interface()
		}
		fieldValue := val.FieldByName(field.Name)
		if vv, err := coerce(value, fieldValue.Interface(), ""); err == nil {
			fieldValue.Set(reflect.ValueOf(vv))
		} else {
			return err
		}
	}
	return nil
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
