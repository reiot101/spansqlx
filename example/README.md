# /example

### Setting up local enviroment spanner-emulator
https://cloud.google.com/spanner/docs/emulator

create gcloud project.
```sh
$ gcloud config configurations create emulator
$ gcloud config set auth/disable_credentials true
$ gcloud config set project sandbox
$ gcloud config set api_endpoint_overrides/spanner http://0.0.0.0:9020/
```

create spanner instance, database.
```sh
$ gcloud config list
[api_endpoint_overrides]
spanner = http://0.0.0.0:9020/
[auth]
disable_credentials = true
[core]
disable_usage_reporting = False
project = sandbox

Your active configuration is: [emulator]

# To switch between the emulator end default configuration, run
$ gcloud config configurations activate [ emulator | default ]

# Create a spanner instance
$ gcloud spanner instances create sandbox --config=emulator-config --description="develop sandbox" --nodes=1

# Show the project spanner instances
$ gcloud spanner instances list --project=sandbox
NAME              DISPLAY_NAME     CONFIG           NODE_COUNT  STATE
sandbox           develop sandbox  emulator-config  1           READY

# Create a database
$ gcloud spanner databases create sandbox --instance sandbox
Creating database...done.

# Show the instance databases
$ gcloud spanner databases list --instance sandbox
NAME        STATE  VERSION_RETENTION_PERIOD  EARLIEST_VERSION_TIME  KMS_KEY_NAME
sandbox  READY
```

use docker compose run spanner-emulator.
```sh
# run the spanner-emulator
$ docker compose up -d

# prune the spanner-emulator
$ docker compose down
```