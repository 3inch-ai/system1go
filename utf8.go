package system1go

import (
	"reflect"
	"strings"
	"unicode/utf8"
)

// Reject invalid source strings before encoding/json can silently replace bytes.
// A depth guard also handles cycles without retaining caller-owned objects.
func validStrings(v reflect.Value, depth int) bool {
	if !v.IsValid() {
		return true
	}
	if depth > 256 {
		return false
	}
	switch v.Kind() {
	case reflect.String:
		return utf8.ValidString(v.String())
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return validStrings(v.Elem(), depth+1)
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if !validStrings(iter.Key(), depth+1) || !validStrings(iter.Value(), depth+1) {
				return false
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !validStrings(v.Index(i), depth+1) {
				return false
			}
		}
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := typ.Field(i)
			if f.PkgPath != "" || strings.Split(f.Tag.Get("json"), ",")[0] == "-" {
				continue
			}
			if !validStrings(v.Field(i), depth+1) {
				return false
			}
		}
	}
	return true
}
