Various utility packages:
* [x] config: weight in the complexity of the viper registry.
* [x] it is just too complex to manage defaults with viper in a deep, hierarchized config
  Try and adopt a struct-driven config with viper unmarshalling
* tracing with NATS: to be tested with a real agent (e.g. jaeger)
* viper patch: to be factorized out in fredbi/go-cli
* pgrepo: to be factorized out in fredbi/pgutils/pgpool
* added support for cluster autodiscovery using headless service: nslookup on headless service
  e.g. flag like '--nats-cluster-headless-service xyz' which overrides '--nats-cluster-routes'
* consumer.process() to add some randomization
* kubernetes deployment:
  - deployment n replicas for 1 producer
  - 1 deployment p replicate for consumer FRPP
  - 1 deployment p' replicate for consumer FRBNP (IBAN prefix)
