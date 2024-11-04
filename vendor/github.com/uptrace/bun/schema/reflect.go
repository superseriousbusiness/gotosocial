package schema

import (
	"database/sql/driver"
	"encoding/json"
	"net"
	"net/netip"
	"reflect"
	"time"
)

var (
	bytesType          = reflect.TypeOf((*[]byte)(nil)).Elem()
	timePtrType        = reflect.TypeOf((*time.Time)(nil))
	timeType           = timePtrType.Elem()
	ipType             = reflect.TypeOf((*net.IP)(nil)).Elem()
	ipNetType          = reflect.TypeOf((*net.IPNet)(nil)).Elem()
	netipPrefixType    = reflect.TypeOf((*netip.Prefix)(nil)).Elem()
	netipAddrType      = reflect.TypeOf((*netip.Addr)(nil)).Elem()
	jsonRawMessageType = reflect.TypeOf((*json.RawMessage)(nil)).Elem()

	driverValuerType  = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	queryAppenderType = reflect.TypeOf((*QueryAppender)(nil)).Elem()
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
)

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func fieldByIndex(v reflect.Value, index []int) (_ reflect.Value, ok bool) {
	if len(index) == 1 {
		return v.Field(index[0]), true
	}

	for i, idx := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return v, false
				}
				v = v.Elem()
			}
		}
		v = v.Field(idx)
	}
	return v, true
}
