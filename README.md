# Machine1 CoreDNS plugin

Provides a [CoreDNS](https://coredns.io/) plugin to allow resolving from
`systemd-machined` machine names to addresses.

## Usage

This plugin provides the directive `machines1` that has the following syntax

```
machines1 <zone> [<ttl>]
```

The `<zone>` argument specifies the domain where the plugin should handle
requests, `<name>.<zone>` queries will resolve to the address of the machine
registered with `<name>`.

The `<ttl>` is an optional argument that sets the TTL of the DNS responses
returned by the plugin. By default, this is set to `1800`.

## Example

```
. {
	machines1 node.internal 3600
}
```

All requests made to subdomains of `node.internal` will be handled through the
plugin and the responses will have a TTL of `3600`.
