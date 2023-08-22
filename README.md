# Darius Calliet August 2023

# Job Proccessing API

This application is designed to enable an end users to Queue up messages that will be consumed in a LIFO manner by our worker process to fulfil our business needs. 

Additionally as we deploy this production system, we've built or incorporated tools into the application ecosystem to support engineering goals and will touch on these throughout:

- Usability

- Observability

- Testability

- Availability

- Security


The application system will make use of:


  - a long lived HTTP(S) server, Server A.


  - a redis database, Server B.


  - a scheduled process, Process C.


  - a monitoring service, Service D. 


  - a logging service, Service E.


## Server A   

Server A will have a long running process baked into a container. It allows clients to use the API via HTTP. See `openapi.yaml` for detailed API structure. Security implications for server A include network access and authorization. The server should be accessible via LB -> firewall. Expected traffic should be served to its configured port. Traffic will come from both public consumers, and machine consumers such as the monitoring Service D (tools like prometheous use polling). Authorization can be implemented using Oauth 2.0 specifications, which would allow us to further secure which clients can preform write and / or read actions via scopes. Doing this would require the implementation of an exchange server but I suggest for v1 we forgo authentication and coorindate with the company if an existing user identity solution exists. Server A should plan to implement a message queue wrapper that sits on a redis database. ex. https://github.com/adjust/rmq

## Server B (Redis)

An underlying redis database would support a persistance and backup of data via a Journaling method, it comes with a default of 2 second intervals. This means on total system failure, up to 2 seconds of data may be lost from memory with the rest backed up on disk. The redis database can easily support duplication, which adds extra confidence to the scale of reads we can support on our queues status.

## Process C

An executable program baked into a container can be configured to run at specific cadences. This process can be setup as a long running process that runs its business logic as bootup defined intervals. Alternatively the process can use systemd on a linux VM, this configuration would be set on deployment time as opposed to bootup time, but the systemd can also be adjusted adhoc by a sysadmin logging into an instance or via a redeployment. This process will need its firewall to allow it to send TCP over a configured port (same as Server B), and depending on the business logic workers are expected to preform, may need to allow other forms of traffic. We will keep it at a minimum for now. This process is expected to make use of a message queue package built on redis and will consume processed jobs. ex. https://github.com/adjust/rmq

## Service D

A monitoring service is imperative for the observability of our system. We can implement a trusted solution such as Prometheus. ex. https://github.com/prometheus/client_golang/. Prometheus has a simple polling mechanism and easy setup, its a trusted tool with a great community. Promethius can support database metrics, and process metrics.

## Service E

Some packages will allow the storing and management of logs on disk, it is likely we have a log aggregation service that wants to receive logs from our respective applications over TCP.

## NOTES ON CLOUD ALTERNATIVES

- SQS can be a replacement for use of redis in our service, our Service A and Service C would need a different usage of the client we'd no longer need redis for this solution.

- DataDog can also likely replace the use of Service D and Service E as DataDog has a large suite of connectors to supplement their large metrics dashboard, it is a pricey enterprise solution but worth it if at that scale!

# Getting Started

## Configuration

This process uses [Viper](https://github.com/spf13/viper) to manage configuration. The process uses `APP_CONFIG_PATH` and `APP_CONFIG_FILENAME` to determine what file to consume for configuration. JSON | YAML | TOML files will all be accepted. `APP_CONFIG_PATH` will default to the working directory `CSC/pkg` and `APP_CONFIG_FILENAME` will default to the current `ENV`. `ENV` will default to the value local.

So with no runtime environment variables running a process `./my_app` will load a configuration file like `CSC/pkg/local.yaml`,`CSC/pkg/local.json`, etc.

With runtime environment variables `ENV=production APP_CONFIG_PATH=CSC/pkg/config ./my_app` will load a configuration file like `CSC/pkg/config/production.yaml`, etc.


## Run Swagger Documentation

- create or update `./CSC/pkg/local.yaml`. An example configuration:
  
    - SWAGGER_PATH: `./CSC`
  
    - SWAGGER_FILENAME: `openapi.yaml`

- `go build -o swagger ./cmd/swagger`

- `CSC_HTTP_PORT=3011 ./swagger`

- Visit `http://localhost:3011/swagger

## Run a Redis Server

Retrieve connection information to an existing redis server **or** for local development while in ./CSC run `docker-compose up redis`.

## Run a Postgres Database

Retrieve connection information to an existing database **or** for local development while in ./CSC run `docker-compose up db`.


Note: the first time these databases are instantiated, a volume will be created that holds schema (in postgres case) and data. To remove completely and rebuild startup scripts run `docker-compose down postgres --volumes`

Note: From here as we move to deployment we'd want to consider a secret management solution, Hashicorp, AWS, and Google all offer cloud solutions. May need to investigate if the company has an existing solution.

Note: If you are more familiar with docker, manage this setup locally as needed.


## Run Server A

- create or update `./CSC/pkg/local.yaml`. An example configuration:
  
    - REDIS_PORT: 6379

    - REDIS_PASSWORD: eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81

    - REDIS_HOSTNAME: localhost
  
    - DB_USERNAME: cscjapi
  
    - DB_PASSWORD: eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81

    - DB_PORT: 5433

    - DB_HOSTNAME: localhost

- `go build -o serverA ./cmd/serverA`

- `CSC_HTTP_PORT=3000 ./serverA`

- Visit `http://localhost:3000/v1/jobs


## Run Process C

- create or update `./CSC/pkg/local.yaml`. An example configuration:
  
    - REDIS_PORT: 6379

    - REDIS_PASSWORD: eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81

    - REDIS_HOSTNAME: localhost
  
    - DB_USERNAME: cscjapi
  
    - DB_PASSWORD: eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81

    - DB_PORT: 5433

    - DB_HOSTNAME: localhost

- `go build -o serverA ./cmd/processC`

- `CSC_CRON_SCHEDULE="*/10 * * * * *" ./processC`

- Visit `http://localhost:3000/v1/jobs
