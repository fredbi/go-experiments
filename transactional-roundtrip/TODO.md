Various utility packages:
* [x] config: weight in the complexity of the viper registry.
* [x] it is just too complex to manage defaults with viper in a deep, hierarchized config. Try and adopt a struct-driven config with viper unmarshalling
* [x] viper patch: to be factorized out in fredbi/go-cli
* tracing with NATS: to be tested with a real agent (e.g. jaeger)
* [x] added support for cluster autodiscovery using headless service: nslookup on headless service
  e.g. flag like '--nats-cluster-headless-service xyz' which overrides '--nats-cluster-routes'
* [x] pgrepo: to be factorized out in fredbi/pgutils/pgpool
* [x] consumer.process() to add some randomization
* [x] comment in message (for tests)
* [x] audit trail of failures (for tests) saved by MessageProcessor
* [x] message injection CLI
* kubernetes deployment:
  - [x] deployment n replicas for 1 producer
  - [x] 1 deployment p replicate for consumer FRPP
  - [x] 1 deployment p' replicate for consumer FRBNP (IBAN prefix)
  - [ ] 1 job for injector
* [x] nats embedded logs not available
* [x] nats embedded monitoring not working
* [x] check headless service resolve w/ selectors: ok
* [x] nslookup of headless service not working
* gooseplus errors: first time (ensure goose version), next: cannot acquire lock
* [x] pgrepo errors: context already done -> more initial probe?
* structured logging for NATS embedded server
* pubsub client as an interface for testability
* version chart for helm releases
