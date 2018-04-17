# CARP

CARP is a "CAS Authentication Reverse Proxy" framework.

## Usage

Configure your environment:

```yaml
cas-url: https://192.168.56.2/cas
service-url: http://192.168.56.1:9090
target-url: http://localhost:8070
skip-ssl-verification: true
port: 9090
principal-header: X-CARP-Authentication
```

Start the server:

```go
package main

func main() {
  flag.Parse()

  configuration, err := ReadConfiguration()
  if err != nil {
     panic(err)
  }

  glog.Infof("start carp %s", Version)

  server, err := NewServer(configuration)
  if err != nil {
	panic(err)
  }

  server.ListenAndServe()
}
```
