# Todo

- Add QoS 1 & 2 support to Client

## Client

- A minimal client implementation that abstracts the underlying transport stream
  by providing a client-like API for subscribing and publishing messages. Will
  handle acknowledgements automatically by using a store. All methods will return
  futures to wait for certain acknowledgements. Does additionally handle ping
  requests as well.

## Service

- Will handle reconnects, multiple brokers, backoff strategy and local message
  caching. Does also resubscribe topics on connection loss and missing session.
