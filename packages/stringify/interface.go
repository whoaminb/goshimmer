package stringify

import (
	"fmt"
	"reflect"
	"strconv"
)

func Interface(value interface{}) string {
	switch value.(type) {
	case bool:
		return Bool(value.(bool))
	case string:
		return String(value.(string))
	case []byte:
		return SliceOfBytes(value.([]byte))
	case int:
		return Int(value.(int))
	case uint64:
		return strconv.FormatUint(value.(uint64), 10)
	case reflect.Value:
		typeCastedValue := value.(reflect.Value)
		switch typeCastedValue.Kind() {
		case reflect.Slice:
			return sliceReflect(typeCastedValue)
		case reflect.String:
			return String(typeCastedValue.String())
		case reflect.Int:
			return Int(int(typeCastedValue.Int()))
		case reflect.Uint8:
			return Int(int(typeCastedValue.Uint()))
		case reflect.Ptr:
			return Interface(typeCastedValue.Interface())
		default:
			panic("undefined reflect type: " + typeCastedValue.Kind().String())
		}
	case fmt.Stringer:
		return value.(fmt.Stringer).String()
	default:
		value := reflect.ValueOf(value)
		switch value.Kind() {
		case reflect.Slice:
			return sliceReflect(value)
		case reflect.Map:
			return mapReflect(value)
		default:
			panic("undefined type: " + value.Kind().String())
		}
	}
}
