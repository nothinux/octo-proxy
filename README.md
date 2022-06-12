# 🐙 Octo-Proxy  
Octo-proxy or `octo` is simple TCP & TLS Proxy with mutual authentication and mirroring/shadowing support.

[![Go Report Card](https://goreportcard.com/badge/github.com/nothinux/octo-proxy)](https://goreportcard.com/report/github.com/nothinux/octo-proxy)  ![test status](https://github.com/nothinux/octo-proxy/actions/workflows/test.yml/badge.svg?branch=master)  

### Feature
- Accept TCP connection and forward/mirror it to TCP
- Accept TCP connection and forward/mirror it to TLS (w/ mTLS)
- Accept TLS (w/ mTLS) connection and forward/mirror it to TCP
- Accept TLS (w/ mTLS) connection and forward/mirror it to TLS (w/ mTLS)
- Reload configuration or certificate without dropping connection
- Expose metrics that can be consumed by prometheus

### Usage
#### Run octo with ad-hoc command
```
octo -listener 127.0.0.1:8080 -target 127.0.0.1:80
```

#### Run Octo as TCP Proxy
``` yaml
// config.yaml
servers:
- name: web-proxy
  listener:
    host: 127.0.0.1
    port: 8080
  target:
    host: 127.0.0.1
    port: 80
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
  target:
    host: 127.0.0.1
    port: 80
```

```
octo -config config.yaml
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
  target:
    host: 127.0.0.1
    port: 80
  mirror:
    host: 172.16.0.1
    port: 80
```

```
octo -config config.yaml
```

See all configuration in [CONFIGURATION.md](https://github.com/nothinux/octo-proxy/tree/master/docs/CONFIGURATION.md)

### LICENSE
[LICENSE](https://github.com/nothinux/octo-proxy/blob/main/LICENSE.md)
