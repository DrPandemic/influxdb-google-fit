# InfluxDB Google Fit
Pulls step counts from Google Fitness and push them to InfluxDB.

## How to setup
### Create Google application
1. Go on https://console.cloud.google.com/cloud-resource-manager and create a new project.
2. On https://console.cloud.google.com/apis/credentials/consent create a new external consent screen. You'll need to add test users during those steps.
3. Once the previous steps are done, create a new "Web application" oauth client on https://console.cloud.google.com/apis/credentials/oauthclient. You should set the appropriate "Authorized redirect URIs".
4. Save the oauth credentials that Google just gave you.

### Running the application
1. Copy `env.template` to `.env` and fill all the values.
2. `go build`
3. `GIN_MODE=release ./influxdb_google_fit`
