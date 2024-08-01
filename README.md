This a small tooling program which streamlines producing to and consuming from EventStoreDB.

# Installation
``` bash
go get ./...
```

# Run
Start eventstore db:
``` bash
cd path/to/eventstoredb
docker-compose up
```

**NOTE:** you might have to run `sudo docker-compose up` on some linux systems

To run the main program:
``` bash
go run src/main.go
```