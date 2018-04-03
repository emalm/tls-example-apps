# TLS Example Apps

This collection of example applications contains a pair of applications designed to communicate directly over HTTP and to verify each other using Cloud Foundry instance-identity credentials.


## Dependencies

- [CF CLI](https://github.com/cloudfoundry/cli/releases), v6.30.0 or later
- CF user with the `network.write` scope, to set network policies


## Deploy to Cloud Foundry

### Initial Setup

Clone this repository and make the frontend and backend binaries:

```
git clone https://github.com/emalm/tls-example-apps.git
cd tls-example-apps
make all
```

Set up a base domain for apps:

```
apps_domain=bosh-lite.com # change for your environment
```

Also, make sure you have logged into your CF deployment and targeted an org and a space to host the apps.

If your CF deployment has enabled [DNS-based platform-native service discovery](https://github.com/cloudfoundry/cf-app-sd-release), choose [Option 1](#option-1-dns-service-discovery) below. Otherwise, choose [Option 2](#option-2-public-service-discovery) for the frontend apps to discover the backend instances through a public route.

### <a name="option-1-dns-service-discovery"></a>Option 1: Use DNS-Based Platform Service Discovery

Push the backend app without a public route:

```
cf push backend -p bin/linux/backend -b binary_buildpack -c './backend' -m 32M -k 32M -i 2 -u process --no-start --no-route
cf set-env backend USE_PLATFORM_SERVICE_DISCOVERY true
```

Map an internal service-discovery route to the backend:

```
backend_hostname=backend
cf map-route backend apps.internal --hostname "$backend_hostname"
```

If another space has already taken the internal route with the `backend` hostname, feel free to use an alternative of your choice, so long as you configure that same choice on the frontend apps below.

Push the 'green' copy of the frontend app and grant it access to the backend app:

```
cf push frontend-green -p bin/linux/frontend -b binary_buildpack -c './frontend' -m 32M -k 32M -i 2 --no-start
cf set-env frontend-green USE_PLATFORM_SERVICE_DISCOVERY true
cf set-env frontend-green BACKEND_DOMAIN "$backend_hostname.apps.internal"

cf add-network-policy frontend-green --destination-app backend --protocol tcp --port 9999
```

Push the 'blue' copy of the frontend app and grant it access to the backend app:

```
cf push frontend-blue -p bin/linux/frontend -b binary_buildpack -c './frontend' -m 32M -k 32M -i 2 --no-start
cf set-env frontend-blue USE_PLATFORM_SERVICE_DISCOVERY true
cf set-env frontend-blue BACKEND_DOMAIN "$backend_hostname.apps.internal"

cf add-network-policy frontend-blue --destination-app backend --protocol tcp --port 9999
```

Skip over Option 2 to [finish configuring the apps and start them](#start-apps).

### <a name="option-2-public-service-discovery"></a>Option 2: Use Public Route for Service Discovery

Push the backend app with its default route:

```
cf push backend -p bin/linux/backend -b binary_buildpack -c './backend' -m 32M -k 32M -i 2 --no-start
```

Push the 'green' copy of the frontend app and grant it access to the backend app:

```
cf push frontend-green -p bin/linux/frontend -b binary_buildpack -c './frontend' -m 32M -k 32M -i 2 --no-start
cf set-env frontend-green BACKEND_DISCOVERY_URL "http://backend.${apps_domain}"

cf add-network-policy frontend-green --destination-app backend --protocol tcp --port 9999
```

Push the 'blue' copy of the frontend app and grant it access to the backend app:

```
cf push frontend-blue -p bin/linux/frontend -b binary_buildpack -c './frontend' -m 32M -k 32M -i 2 --no-start
cf set-env frontend-blue BACKEND_DISCOVERY_URL "http://backend.${apps_domain}"

cf add-network-policy frontend-blue --destination-app backend --protocol tcp --port 9999
```


### <a name="start-apps"></a>Finish Configuration and Start the Apps

Configure the backend app to authorize only the 'green' frontend app:

```
FRONTEND_GREEN_APP_GUID=$(cf app frontend-green --guid)
FRONTEND_BLUE_APP_GUID=$(cf app frontend-blue --guid)

cf set-env backend AUTHORIZED_APP_GUIDS "[\"$FRONTEND_GREEN_APP_GUID\"]"
```

Start the apps:

```
cf start backend
cf start frontend-green
cf start frontend-blue
```

Make requests to the frontend apps:

```
curl https://frontend-green.${apps_domain}
curl https://frontend-blue.${apps_domain}
```

The 'green' frontend app will report success, and the 'blue' one will report failure.

Reconfigure the backend to authorize the 'blue' frontend instead:

```
cf set-env backend AUTHORIZED_APP_GUIDS "[\"$FRONTEND_BLUE_APP_GUID\"]"
cf restart backend
```

Now requests to the 'blue' frontend will succeed, and those to the 'green' frontend will fail.


## CF Deployment Configuration

The CF deployment must be configured to use container networking and to enable the Diego cells to generate instance-identity credentials. Version [v0.31.0 of cf-deployment](https://github.com/cloudfoundry/cf-deployment/tree/v0.31.0) with the [enable-instance-identity-credentials](https://github.com/cloudfoundry/cf-deployment/blob/v0.31.0/operations/experimental/enable-instance-identity-credentials.yml) operations file will be configured this way.


## Local Development

On Mac OS X:

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

curl http://127.0.0.1:8081
```
