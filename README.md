# CARP

CAS Authentication Reverse Proxy

## Requirements

* [Go](https://golang.org/) >= 1.10
* [Dep](https://golang.github.io/dep/)

## PoC with Nexus

* start Cloudogu EcoSystem
* finish Setup and install at least cas and usermgt
* enable development mode
```bash
etcdctl set /config/_global/stage development
cesapp stop cas
cesapp start cas
```
* start nexus
```bash
docker-compose up -d
```
* open Nexus at http://localhost:8081
* sign in with admin and admin123
* enable "RUT Auth Realm" at Security->Realms
* Add "RUT Auth" Capability with X-CARP-Authentication as Header
* Checkout CARP
```bash
mkdir -p ${GOPATH}/src/github.com/cloudogu
cd ${GOPATH}/src/github.com/cloudogu
git clone git@github.com:cloudogu/carp.git
cd carp
```
* Build CARP
```bash
dep ensure
go build
```
* Run CARP
```bash
./carp
```
* Test Nexus with Browser and Maven at http://192.168.56.1:9090
