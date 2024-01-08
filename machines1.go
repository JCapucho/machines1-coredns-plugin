package machines1

import (
	"context"
	"errors"

	"golang.org/x/sys/unix"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/godbus/dbus/v5"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("machines1")

type Machines1 struct {
	Next plugin.Handler

	obj dbus.BusObject

	ttl  uint32
	from string
}

func (m Machines1) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()

	zone := plugin.Zones([]string{m.from}).Matches(qname)
	if zone == "" {
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	if len(qname) == len(zone) {
		msg.Rcode = dns.RcodeNameError
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}
	host := qname[:len(qname)-len(zone)-1]

	call := m.obj.CallWithContext(
		ctx, "org.freedesktop.machine1.Manager.GetMachineAddresses", 0, host)

	if call.Err != nil {
		var e dbus.Error
		if errors.As(call.Err, &e) && e.Name == "org.freedesktop.machine1.NoSuchMachine" {
			log.Debug("No machine running")
			msg.Rcode = dns.RcodeNameError
		} else {
			log.Error("Error: ", call.Err)
			return dns.RcodeServerFailure, plugin.Error("machines1", call.Err)
		}
	} else {
		for _, resInt := range call.Body {
			res := resInt.([][]interface{})

			for _, address := range res {
				family := address[0].(int32)
				address := address[1].([]byte)
				switch family {
				case unix.AF_INET:
					if state.QType() == dns.TypeA {
						r := new(dns.A)
						r.Hdr = dns.RR_Header{Name: qname, Ttl: m.ttl, Class: dns.ClassINET, Rrtype: dns.TypeA}
						r.A = address
						msg.Answer = append(msg.Answer, r)
					}
				case unix.AF_INET6:
					if state.QType() == dns.TypeAAAA {
						r := new(dns.AAAA)
						r.Hdr = dns.RR_Header{Name: qname, Ttl: m.ttl, Class: dns.ClassINET, Rrtype: dns.TypeAAAA}
						r.AAAA = address
						msg.Answer = append(msg.Answer, r)
					}
				default:
					log.Warning("Unknown address family: ", family, " (address = ", address, ")")
				}
			}
		}
	}

	w.WriteMsg(msg)
	return dns.RcodeSuccess, nil
}

func (m Machines1) Name() string { return "machines1" }
