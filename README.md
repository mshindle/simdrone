# simdrone

## requirements

Simdrone needs a bus to work correctly with RabbitMQ as the most popular choice.
For running locally, the following command will get you quickly set up.

```shell
docker run -d --hostname rabbit --name rabbitmq -p 8080:15672 -p 4369:4369 -p 5672:5672 rabbitmq:3-management
```

## cmds / handler

## events

