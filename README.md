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

