# TLS Example Apps

This collection of example applications contains a pair of applications designed to communicate directly over HTTP and to verify each other using Cloud Foundry instance-identity credentials.

## Local Development

```
make all

CERT_RELOAD_INTERVAL=10s \
CF_INSTANCE_GUID=backend-1 \
CF_INSTANCE_INTERNAL_IP=127.0.0.1 \
CF_INSTANCE_CERT=creds/server.crt \
CF_INSTANCE_KEY=creds/server.key \
CA_CERT_FILE=creds/ca.crt \
./bin/darwin/backend

PORT=8081 \
CERT_RELOAD_INTERVAL=10s \
CF_INSTANCE_CERT=creds/client.crt \
CF_INSTANCE_KEY=creds/client.key \
CA_CERT_FILE=creds/ca.crt \
./bin/darwin/frontend

curl http://127.0.0.1:8081?ip=127.0.0.1
```

