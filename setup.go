package machines1

import (
	"fmt"
	"strconv"

	"github.com/coredns/caddy"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/godbus/dbus/v5"
)

func init() { plugin.Register("machines1", setup) }

func setup(c *caddy.Controller) error {
	c.Next()

	from := ""
	ttlString := ""
	var ttl uint64 = 1800

	if !c.Args(&from) {
		return plugin.Error("machines1", c.ArgErr())
	}
	if c.Args(&ttlString) {
		var err error
		ttl, err = strconv.ParseUint(ttlString, 10, 32)
		if err != nil {
			return plugin.Error("machines1", fmt.Errorf("failed to parse ttl '%s': %s", ttlString, err))
		}
	}
	if c.NextArg() {
		return plugin.Error("machines1", c.ArgErr())
	}

	normalized := plugin.Host(from).NormalizeExact()
	if len(normalized) == 0 {
		return plugin.Error("machines1", fmt.Errorf("unable to normalize '%s'", from))
	}
	from = normalized[0]

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return plugin.Error("machines1", err)
	}

	obj := conn.Object("org.freedesktop.machine1", "/org/freedesktop/machine1")

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Machines1{Next: next, obj: obj, ttl: uint32(ttl), from: from}
	})

	c.OnShutdown(func() error {
		return conn.Close()
	})

	return nil
}
