# Nexus

### Base URL

POST http://localhost:9090/service/extdirect
```json
{"action":"capability_Capability","method":"create","data":[{"id":"","typeId":"baseurl","notes":"","enabled":true,"properties":{"url":"http://localhost:9090/"}}],"type":"rpc","tid":13}
```

### Enable realm

POST http://192.168.56.1:8081/service/extdirect
```json
{"action":"coreui_RealmSettings","method":"update","data":[{"realms":["NexusAuthenticatingRealm","NexusAuthorizingRealm","rutauth-realm"]}],"type":"rpc","tid":13}
```

### Create Capability

POST http://192.168.56.1:8081/service/extdirect
```json
{"action":"capability_Capability","method":"create","data":[{"id":"","typeId":"rutauth","notes":"","enabled":true,"properties":{"httpHeader":"X-CARP-Authentication"}}],"type":"rpc","tid":19}
```

