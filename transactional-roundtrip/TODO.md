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
  - 1 job for injector
* gooseplus errors: first time (ensure goose version), next: cannot acquire lock
* pgrepo errors: context already done
