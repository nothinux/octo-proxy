# 🐙 Octo-Proxy  
Octo-proxy or `octo` is simple TCP & TLS Proxy with mutual authentication and traffic mirroring/shadowing support.

[![Go Report Card](https://goreportcard.com/badge/github.com/nothinux/octo-proxy)](https://goreportcard.com/report/github.com/nothinux/octo-proxy)  ![test status](https://github.com/nothinux/octo-proxy/actions/workflows/test.yml/badge.svg?branch=main)  [![codecov](https://codecov.io/gh/nothinux/octo-proxy/branch/main/graph/badge.svg?token=HBRTW7DX0K)](https://codecov.io/gh/nothinux/octo-proxy)

### Feature
- Accept TCP connection and forward/mirror it to TCP
- Accept TCP connection and forward/mirror it to TLS (w/ mTLS)
- Accept TLS (w/ mTLS) connection and forward/mirror it to TCP
- Accept TLS (w/ mTLS) connection and forward/mirror it to TLS (w/ mTLS)
- Support for multiple targets, accessed in random order (load balancer)
- Reload configuration or certificate without dropping connection
- Expose metrics that can be consumed by prometheus

### Usage
#### Run octo with ad-hoc command
```
octo-proxy -listener 127.0.0.1:8080 -target 127.0.0.1:80
```

Run with `-debug` to get a more verbose log output.

#### Run Octo as TCP Proxy with metrics on port 9123
``` yaml
// config.yaml
servers:
- name: web-proxy
  listener:
    host: 127.0.0.1
    port: 8080
  targets:
    - host: 127.0.0.1
      port: 80
    - host: 127.0.0.1
      port: 81

metrics:
  host: 0.0.0.0
  port: 9123
```

```
octo -config config.yaml
```

#### Run Octo as TLS Proxy w/ mTLS
``` yaml
// config.yaml
servers:
- name: web-proxy
  listener:
    host: 0.0.0.0
    port: 8080
    tlsConfig:
      mode: mutual
      caCert: /tmp/ca-cert.pem
      cert: /tmp/cert.pem
      key: /tmp/cert-key.pem
  targets:
    - host: 127.0.0.1
      port: 80

metrics:
  host: 0.0.0.0
  port: 9123
```

```
octo-proxy -config config.yaml
```

#### Run Octo as TLS Proxy and Mirror traffic to other backend
``` yaml
// config.yaml
servers:
- name: web-proxy
  listener:
    host: 0.0.0.0
    port: 8080
    tlsConfig:
      mode: simple
      cert: /tmp/cert.pem
      key: /tmp/cert-key.pem
  targets:
    - host: 127.0.0.1
      port: 80
  mirror:
    host: 172.16.0.1
    port: 80
```

```
octo-proxy -config config.yaml
```

See all configuration in [CONFIGURATION.md](https://github.com/nothinux/octo-proxy/tree/master/docs/CONFIGURATION.md)

### Reloading Octo-proxy
After changing configuration or certificates, send signal `SIGUSR1` or `SIGUSR2` to `octo-proxy` process. Configuration will be reloaded if the configuration is valid.

Octo-proxy use `SO_REUSEPORT` to binding the listener, so every reload triggered octo-proxy will create new listener and drop old listener after new listener created, by using this approach octo-proxy can minimize dropped connection when reload triggered.

### Monitoring
Metrics are configured through the `metrics` section in the config file and are served under the `/metrics` path of the configured host and port.

### LICENSE
[LICENSE](https://github.com/nothinux/octo-proxy/blob/main/LICENSE.md)
