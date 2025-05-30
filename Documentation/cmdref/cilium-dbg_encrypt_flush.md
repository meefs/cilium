<!-- This file was autogenerated via cilium-dbg cmdref, do not edit manually-->

## cilium-dbg encrypt flush

Flushes the current IPsec state

### Synopsis

Will cause a short connectivity disruption

```
cilium-dbg encrypt flush [flags]
```

### Options

```
  -f, --force            Skip confirmation
  -h, --help             help for flush
      --node-id string   Only delete states and policies with this node ID. Decimal or hexadecimal (0x) format. If multiple filters are used, they all apply
  -o, --output string    json| yaml| jsonpath='{}'
      --spi uint8        Only delete states and policies with this SPI. If multiple filters are used, they all apply
      --stale            Delete stale states and policies based on the current node ID map content
```

### Options inherited from parent commands

```
      --config string        Config file (default is $HOME/.cilium.yaml)
  -D, --debug                Enable debug messages
  -H, --host string          URI to server-side API
      --log-driver strings   Logging endpoints to use (example: syslog)
      --log-opt map          Log driver options (example: format=json)
```

### SEE ALSO

* [cilium-dbg encrypt](cilium-dbg_encrypt.md)	 - Manage transparent encryption

