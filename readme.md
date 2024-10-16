# Go Programming Exercise - To-Do App



## Quickstart

Current phase of development: 2

To see a full list of flags available to the CLI use `go run . -h`

### Windows

1. Build the application :<br> 
`go build -o todo-app.exe` 
2. Start the Server in it's own terminal:<br>
`start todo-app.exe --start-server --mode=in-mem`
3. Use the CLI to make a test request to the server:<br>
`go run . --post --id=3fa85f64-5717-4562-b3fc-2c963f66afa6 --title=test --priority=high`

### Linux

1. Start the server in the background <br>
`go run . --start-server --mode=in-mem &`
2. Use the CLI to make a test request to the server:<br>
`go run . --post --id=3fa85f64-5717-4562-b3fc-2c963f66afa6 --title=test --priority=high`

## API

<pr>The API spec can found at http://localhost:8081/swagger-ui</pr>

## Key Design Decisions

### Phase 3

- Approaching the update to the api as a breaking change.
- Treating updates as specific to a user id. User "123" is not able to access or update items from their ownership to another user. Nor can they access items belonging to another user.