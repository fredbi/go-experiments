# go-experiments
Personal musings with golang

## transactional-roundtrip

An experiment with deploying a cluster of embedded NATS servers that
implement at-least-once delivery over a fire-and-forget messaging network.

The example server simulates bank transfers routed over a cluster of _collaborating_ nodes,
with a guaranteed eventual known outcome.

Consistency is achieved by relentless idempotent deliveries from the endpoints, with only one transactional store (postgres).

The protocol boils down to 3 phases: send -> respond -> confirm response.

The example builds self-contained docker images and a k8s setup,
with a test injector for POC. Local testing with `minikube`.

* Keywords: API, publish-subscribe, at-least-once
* Stack: `NATS`, `postgres`, `docker`, `kubernetes`

**Thoughts & could-be-done**:

1. Investigate: if we add the possibility of a give-up (e.g. after some time or a max number of
   attempts), what do we need to change to the protocol to keep consistency? I.e. introduce a "Cancel" operation to 
   revert any given-up transaction (aka "saga"-style transactional guarantees).
2. Investigate: if we add a guaranteed transport
   with ACK, e.g. NATS jetstream  - compares to kafka, rabbitMQ, Google PubSub - how does this translate in terms of:
   * simplifying things? (atm I think it doesn't really simplify anything)
   * achieving better resilience in noisy environments (Q: failure rate threshold at which point adopting a guaranteed transport
     ends up with less redeliveries and network gossip)
3. Partipant to the mesh run independant databases, yet global consistency. In the demo, "consumer" nodes only
   use the DB to track confirmed transfers and only "producer" nodes really govern the persistent state.
4. Investigate JetStream: how about replacing the relational DB by the distributed object store embedded with Jetstream?
   (e.g. a consistent, RAFT-based distributed KV store).
5. If we assume that "consumer" nodes are mere proxies to remote participants, with some extra delays & failures, explore the new appropriate protocol.

### Non-goals

* Nothing about security, untrusted nodes or anything of the sort can be found here

### Micro-experiments within the experiment**:

* NATS tracing (uses [go-trace](https://github.com/fredbi/go-trace/tree/master/tracer))
* In-deployment, embedded DB migrations (no need for a k8s job to coordinate the deployment) (uses [gooseplus](https://github.com/fredbi/gooseplus))
* Multi-modules configuration, 12-factor app made simple (uses [go-cli/config](https://github.com/fredbi/go-cli/tree/master/config))
* Advanced CLI with cobra (e.g. help topics, dependency injection) (uses [go-cli/cli](https://github.com/fredbi/go-cli/tree/master/cli))
* Streamed DB queries using generic iterators  (uses [go-patterns/iterators](https://github.com/fredbi/go-patterns/tree/master/iterators)
* Playing with decimals (uses cockroachdb's [apd.Decimal](https://github.com/cockroachdb/apd))
* Interesting musings from other people:
  - https://github.com/ripienaar/nats-kv-leader-elect
