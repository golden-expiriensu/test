## Description

On startup application declares `input-A` and `output-A` queues and starts listen on `input-A`.
Upon delivery service processes it (basically does `{"result": " + string(input) + "}`).
Processed message is being published to default exchange with routing key `output-A`.

### Black box
Client publishes 'hello' to `input-A` and receives '{"result": "hello"}' from `output-A`.


## How to

### Lint
Format the code, fix dependencies and validate the project.
```bash
make lint
```

### Run
Bootstrap RabbitMQ, run the service and start processing messages received from the queue.
```bash
make
```

## Objective

Add support for the second stream of data using `input-B` and `output-B` queues.

### Black box
1. Client publishes 'hello' to `input-A` and receives '{"result": "hello"}' from `output-A`.
2. Client publishes 'hello' to `input-B` and receives '{"result": "hello"}' from `output-B`.


#### Running the Black Box

Support for B has been added

Create streams C
  ```bash
  curl -X POST -d '{"name": "C"}' http://localhost:8080/create
  ```

Publish streams A, B, C
  ```bash
  curl -X PUT -d '{"message": "Hello A"}' http://localhost:8080/publish/input-A
  curl -X PUT -d '{"message": "Hello B"}' http://localhost:8080/publish/input-B
  curl -X PUT -d '{"message": "Hello C"}' http://localhost:8080/publish/input-C
  ```

Consume from streams A, B or C
  ```bash
  curl http://localhost:8080/consume/output-A
  curl http://localhost:8080/consume/output-B
  curl http://localhost:8080/consume/output-C
  ```

# Bonus feature tip:
  publish multiple messages into a particular stream and consume them back in array of json responses

