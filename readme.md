# Go Programming Exercise - To-Do App

## Current phase of development: 3

## Quickstart

The project has two `main` packages, the [CLI] & the [Server].

The ToDo [CLI] acts as a command line client application that can make requests to the ToDo [Server].

The ToDo [Server] hosts the endpoints for the [V1 API], [V2 API] and the client appilication (add more info here)

### Server application

To get started with the Server Application, see [server docs]

### CLI

Before running the [CLI] it's important to start the [Server] application, as it makes requests via the REST API. 

When you're ready to get started, see [CLI docs]

[CLI]: cli/cli.go
[Server]: to-do-server/to-do-server.go
[CLI docs]: cli/readme.md
[server docs]: to-do-server/readme.md
[V1 API]: to-do-server/api-specs/to-do-app-api-v1.yaml
[V2 API]: to-do-server/api-specs/to-do-app-api-v2.yaml
