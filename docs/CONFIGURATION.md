# Configuration

## Servers
| Field    | Type          | Description                     | Required |
| -------- | ------------- | ------------------------------- | -------- |
| servers  | [Server[]](#server)      | A list of proxy exposed by octo | yes      |


## Server
| Field    | Type             | Description   | Required |
| -------- | ---------------- | ------------- | -------- |
| name     | `<string>`       | Name of proxy | no       |
| listener | [`Hostconfig`](#hostconfig) | Set of listener related configuration. All of the incoming request to octo-proxy will be handled by this listener.            | yes      |
| targets  | [`Hostconfig[]`](#hostconfig) | Set of target related configurations. These targets are backends which octo-proxy will forward all incoming traffic accepted by the listener.            | yes      |
| mirror   | [`Hostconfig`](#hostconfig)  | Set of mirror related configuration. If this configuration is enabled, all incoming requests will also be forwarded to this mirror. Unlike the `target`, in a `mirror` setup, we implement 'fire and forget,' where every request is only forwarded, and the response is ignored.          | no       |

## Hostconfig
| Field     | Type          | Description                     | Required |
| --------- | ------------- | ------------------------------- | -------- |
| host      | `<string>`    | On the `listener`, this is host to which the listener will be listen, and on `target` and `mirror` this is the host of the backend to which the request will be forwarded | yes      |
| port      | `<string>`    | On the `listener`, this is port to which the listener will be bind, and on `target` and `mirror` this is the port of the backend to which the request will be forwarded | yes      |
| connection   | [`connectionConfig`](#connectionConfig)    | set timeout/deadline (in seconds) for every connections, default 300 seconds. A value of 0 will disable deadlines on connections | no      |
| tls       | [`tlsConfig`](#tlsconfig)   | set tls configuration if the host is using tls | no      |


## connectionConfig
| Field    | Type          | Description                     | Required |
| -------- | ------------- | ------------------------------- | -------- |
| timeout  | `<string>`    | Set timeout or deadline for every connection, you can setthe unit in milliseconds with `ms` or seconds with `s`. the default value is `300 seconds``. A value of 0 will disable deadlines on connections                 | no       |

## tlsConfig
| Field    | Type          | Description                     | Required |
| -------- | ------------- | ------------------------------- | -------- |
| mode     | [`tlsMode`](#tlsmode)       | Set mode of tls                 | yes      |
| caCert   | `<string>`    | The path to the CA Certificate file, the CA will be used to verify server/client certificate. In `simple` mode, this option allows you to verify the server certificate if the CA is not stored in the trust store. For `mutual` mode this option is `REQUIRED` to verify client certificate                | yes      |
| cert     | `<string>`    | The path to the Certificate file, this option need to be set if the mode is `mutual` to authenticate/verify the server/client.                   | yes      |
| key      | `<string>`    | The path to the Private key file, this option need to be set if the mode is `mutual`.                  | yes      |
| sni      | `<string>`    | Set SNI during TLS handshake                  | no      |
| crl      | `<string>`    | The path to CRL file, If the CRL is configured, the server/client will verify the peer's certificate against the CRL     | no      |

## tlsMode
| Field     | Type          | Description                     |
| --------- | ------------- | ------------------------------- 
| `simple`  | `<string>`    | Use this option to enable simple TLS. In this mode octo will only verify the server identity. Required option `mode: simple` and `caCert` if the root CA is not stored in the trust store |
| `mutual`  | `<string>`    | Use this option to enable mutual TLS (mTLS). In this mode, the server and client will verify each other. Required option `mode: mutual`, `caCert`, `cert`, and `key`. |

> Currently, in mutual mode octo-proxy will only verify the ip address of it's client and try to match it with ip sans in certificate. In the future we will adding more alternative names verification.

## Metrics
| Field    | Type          | Description                     | Required |
| -------- | ------------- | ------------------------------- | -------- |
| metrics  | [HostConfig]  | Configures the host and port for the metrics server, currently doesn't support tls settings | no       |
