# gostack

Integration test framework for OpenStack-to-OpenStack and VMware-to-OpenStack migration. Provides a fake OpenStack API server (Keystone, Nova, Neutron, Cinder, Glance) for Ansible-based integration tests.

## Usage

### Run the fake server

```bash
go run ./cmd/fake-openstack
# or
make build && ./bin/fake-openstack
```

### Configuration (viper + cobra)

- `--port`, `-p` (default: 5000)
- `--bind`, `-b` (default: 127.0.0.1)
- `--pid-file` (default: /tmp/fake_os_server.pid)
- `--log-level` (debug, info, warn, error)
- `--base-url` (default: http://bind:port)

Environment: `GOSTACK_PORT`, `GOSTACK_BIND`, etc.

### As a Go package

```go
import "github.com/os-migrate/gostack/pkg/gostack"

opts := gostack.DefaultOptions()
opts.Port = 5000
fs := gostack.NewFakeServer(opts)
defer fs.Close()
// Server runs at fs.URL
select {} // or run your tests
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI
- [viper](https://github.com/spf13/viper) - Config
- [logrus](https://github.com/sirupsen/logrus) - Logging
- [testify](https://github.com/stretchr/testify) - Testing

## License

Apache-2.0
