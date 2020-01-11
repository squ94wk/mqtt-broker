# mqtt-broker

A (eventually) fully MQTT 5.0 compliant broker.

## Features
MQTT v5.0 compliance
* Connect => Connack
* Disconnect
* Session takeover
* Subscribe => Suback
* Publish => delivery

Yet to be implemented:
* Error handling
* Keep Alive & session expiry
* Unsubscribe
* Ping
* Authentication
* Proper error handling
* QoS 1
* QoS 2
* Quotas
* Message retention
* Will messages
* Topic aliases
* Shared subscriptions
* Large Payload handling
* Security
    * Traffic control
    * Blacklisting
* Tests for MQTT compliance

Ops
* Configuration
* Observability
    * Tracing
    * Metrics
* Admin panel
* Docker image

Replication
* Synchronization: Sessions, Message broadcasting
* Redundancy
* Fail-Over
* Storage options
    * in memory
    * distributed FS
    * key-value store
* Health check
* K8s manifests

Performance
* Benchmarking
* Stress tests
