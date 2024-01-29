# CQRS

CQRS splits your application (and even the database in some cases) into two different paths: **Commands** and **Queries**.

## Command side

Every operation that can trigger an side effect on the server must pass through the CQRS "command side". I like to put the `Handlers` (commands handlers and events handlers) inside the application layer because their goals are almost the
same: orchestrate domain operations (also usually using infrastructure services).

![command side](docs/images/command_side.jpg)

[//]: # (![command side]&#40;docs/images/command_side_with_events.jpg&#41;)

## Query side

Pretty straight forward, the controller receives the request, calls the related query repo and returns a DTO (defined on infrastructure layer itself).

![query side](docs/images/query_side.jpg)
