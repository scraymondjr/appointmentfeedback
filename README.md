# Patient Appointment Survey

Service for holding data pertaining to patients and their appointments 
and allow patients to respond to surveys about their experience.

## Commands

Main CLI entry and help for available commands:
```shell
go run cmd/cli/main.go help
```

#### Interact with system as a patient:
```shell
go run cmd/cli/main.go patient ...
```

## Neo4j

Start local container instance:
```
docker run --name testneo4j -p7474:7474 -p7687:7687 --rm -d -v $HOME/neo4j/data:/data -v $HOME/neo4j/logs:/logs -v $HOME/neo4j/import:/var/lib/neo4j/import -v $HOME/neo4j/plugins:/plugins --env NEO4J_AUTH=none neo4j:latest
```

[Browser REPL](http://localhost:7474/browser/)