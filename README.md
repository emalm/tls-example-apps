# TLS Example Apps

This collection of example applications contains a pair of applications designed to communicate directly over HTTP and to verify each other using Cloud Foundry instance-identity credentials.

## Local Development

### Backend app

```
make backend

CF_INSTANCE_CERT=creds/server.crt \
CF_INSTANCE_KEY=creds/server.key \
./bin/darwin/backend
```

