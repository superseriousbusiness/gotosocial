package mangler

import (
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"time"
	_ "unsafe"
)

func init() {
	// Register standard library performant manglers.
	Register(reflect.TypeOf(net.IPAddr{}), mangle_ipaddr)
	Register(reflect.TypeOf(&net.IPAddr{}), mangle_ipaddr_ptr)
	Register(reflect.TypeOf(netip.Addr{}), mangle_addr)
	Register(reflect.TypeOf(&netip.Addr{}), mangle_addr_ptr)
	Register(reflect.TypeOf(netip.AddrPort{}), mangle_addrport)
	Register(reflect.TypeOf(&netip.AddrPort{}), mangle_addrport_ptr)
	Register(reflect.TypeOf(time.Time{}), mangle_time)
	Register(reflect.TypeOf(&time.Time{}), mangle_time_ptr)
	Register(reflect.TypeOf(url.URL{}), mangle_url)
	Register(reflect.TypeOf(&url.URL{}), mangle_url_ptr)
}

//go:linkname time_sec time.(*Time).sec
func time_sec(*time.Time) int64

//go:linkname time_nsec time.(*Time).nsec
func time_nsec(*time.Time) int32

//go:linkname time_stripMono time.(*Time).stripMono
func time_stripMono(*time.Time)

func mangle_url(buf []byte, a any) []byte {
	u := (*url.URL)(iface_value(a))
	return append(buf, u.String()...)
}

func mangle_url_ptr(buf []byte, a any) []byte {
	if ptr := (*url.URL)(iface_value(a)); ptr != nil {
		s := ptr.String()
		buf = append(buf, '1')
		return append(buf, s...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_time(buf []byte, a any) []byte {
	t := *(*time.Time)(iface_value(a))
	time_stripMono(&t) // force non-monotonic time value.
	buf = append_uint64(buf, uint64(time_sec(&t)))
	buf = append_uint32(buf, uint32(time_nsec(&t)))
	return buf
}

func mangle_time_ptr(buf []byte, a any) []byte {
	if ptr := (*time.Time)(iface_value(a)); ptr != nil {
		t := *ptr
		buf = append(buf, '1')
		time_stripMono(&t) // force non-monotonic time value.
		buf = append_uint64(buf, uint64(time_sec(&t)))
		buf = append_uint32(buf, uint32(time_nsec(&t)))
		return buf
	}
	buf = append(buf, '0')
	return buf
}

func mangle_ipaddr(buf []byte, a any) []byte {
	i := *(*net.IPAddr)(iface_value(a))
	buf = append(buf, i.IP...)
	buf = append(buf, i.Zone...)
	return buf
}

func mangle_ipaddr_ptr(buf []byte, a any) []byte {
	if ptr := (*net.IPAddr)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		buf = append(buf, ptr.IP...)
		buf = append(buf, ptr.Zone...)
		return buf
	}
	buf = append(buf, '0')
	return buf
}

func mangle_addr(buf []byte, a any) []byte {
	i := (*netip.Addr)(iface_value(a))
	b, _ := i.MarshalBinary()
	return append(buf, b...)
}

func mangle_addr_ptr(buf []byte, a any) []byte {
	if ptr := (*netip.Addr)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		b, _ := ptr.MarshalBinary()
		return append(buf, b...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_addrport(buf []byte, a any) []byte {
	i := (*netip.AddrPort)(iface_value(a))
	b, _ := i.MarshalBinary()
	return append(buf, b...)
}

func mangle_addrport_ptr(buf []byte, a any) []byte {
	if ptr := (*netip.AddrPort)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		b, _ := ptr.MarshalBinary()
		return append(buf, b...)
	}
	buf = append(buf, '0')
	return buf
}
