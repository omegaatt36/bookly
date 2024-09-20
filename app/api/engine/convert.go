package engine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/omegaatt36/bookly/app"
)

func condConvert(payload string, val any) error {
	typKind := reflect.Indirect(reflect.ValueOf(val)).Type().Kind()
	rTyp := reflect.ValueOf(val).Elem()
	if !rTyp.CanSet() {
		return app.ParamError(fmt.Errorf("can't set type %v", rTyp.Type()))
	}
	switch typKind {
	case reflect.String:
		rTyp.Set(reflect.ValueOf(payload).Convert(rTyp.Type()))
	case reflect.Uint, reflect.Int:
		i, err := strconv.Atoi(payload)
		if err != nil {
			return app.ParamError(fmt.Errorf("parse int('%v') failed: %v", payload, err))
		}
		rTyp.Set(reflect.ValueOf(i).Convert(rTyp.Type()))
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(payload)
		if err != nil {
			return app.ParamError(fmt.Errorf("parse bool('%v') failed: %v", payload, err))
		}
		rTyp.Set(reflect.ValueOf(boolVal).Convert(rTyp.Type()))
	case reflect.Struct:
		if isTimeStruct(val) {
			parsedTime, err := time.ParseInLocation(time.RFC3339, payload, time.Local)
			if err != nil {
				return app.ParamError(fmt.Errorf("parse time rfc3339('%v') failed: %v", payload, err))
			}
			rTyp.Set(reflect.ValueOf(parsedTime).Convert(rTyp.Type()))
		} else {
			return app.ParamError(fmt.Errorf("unsupported struct type(%v)", typKind))
		}
	case reflect.Slice:
		sliceKind := reflect.TypeOf(val).Elem().Elem().Kind()
		switch sliceKind {
		case reflect.Uint:
			ss := strings.Split(payload, ",")
			arr := make([]uint, len(ss))
			for i, s := range ss {
				v, err := strconv.Atoi(s)
				if err != nil {
					return app.ParamError(fmt.Errorf("parse uint('%v') failed: %v", s, err))
				}

				arr[i] = uint(v)
			}
			rTyp.Set(reflect.ValueOf(arr).Convert(rTyp.Type()))
		default:
			return app.ParamError(fmt.Errorf("unsupported slice type(%v)", sliceKind))
		}
	default:
		return app.ParamError(fmt.Errorf("unsupported type(%v)", typKind))
	}
	return nil
}

func isTimeStruct(x any) bool {
	typ := reflect.TypeOf(x)
	return typ.String() == "*time.Time"
}
