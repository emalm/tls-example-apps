# TLS Example Apps

This collection of example applications contains a pair of applications designed to communicate directly over HTTP and to verify each other using Cloud Foundry instance-identity credentials.

## Deploy to Cloud Foundry

```
make all

apps_domain=bosh-lite.com # change for your environment

# push backend app

cf push backend -p bin/linux/backend -b binary_buildpack -c './backend' --no-start

# push 'green' frontend app and allow to access backend

cf push frontend-green -p bin/linux/frontend -b binary_buildpack -c './frontend' --no-start
cf set-env frontend-green BACKEND_DISCOVERY_URL "http://backend.${apps_domain}"

cf add-network-policy frontend-green --destination-app backend --protocol tcp --port 9999


# push 'blue' frontend app and allow to access backend

cf push frontend-blue -p bin/linux/frontend -b binary_buildpack -c './frontend' --no-start
cf set-env frontend-blue BACKEND_DISCOVERY_URL "http://backend.${apps_domain}"

cf add-network-policy frontend-blue --destination-app backend --protocol tcp --port 9999


# configure backend app to authorize frontend app

FRONTEND_GREEN_APP_GUID=$(cf app frontend-green --guid)
FRONTEND_BLUE_APP_GUID=$(cf app frontend-blue --guid)

cf set-env backend AUTHORIZED_APP_GUIDS "[\"$FRONTEND_GREEN_APP_GUID\"]"
cf set-env backend CERT_RELOAD_INTERVAL 5s


# start the apps

cf start backend

cf start frontend-green
cf start frontend-blue


# make requests to the front ends

curl https://frontend-green.${apps_domain}
curl https://frontend-blue.${apps_domain}
```


## Local Development

```
make all

CERT_RELOAD_INTERVAL=10s \
CF_INSTANCE_GUID=backend-1 \
CF_INSTANCE_INTERNAL_IP=127.0.0.1 \
CF_INSTANCE_CERT=creds/server.crt \
CF_INSTANCE_KEY=creds/server.key \
CA_CERT_FILE=creds/ca.crt \
./bin/darwin/backend/backend

PORT=8081 \
CERT_RELOAD_INTERVAL=10s \
CF_INSTANCE_CERT=creds/client.crt \
CF_INSTANCE_KEY=creds/client.key \
CA_CERT_FILE=creds/ca.crt \
./bin/darwin/frontend/frontend

curl http://127.0.0.1:8081?ip=127.0.0.1
```

