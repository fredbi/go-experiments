#  nats-bank

I want to experiment with NATS and deploy a transactional round-trip message processing pattern.

The example server roughly simulates bank transfers routed over a cluster of _collaborating_ nodes,
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

### Micro-experiments within the experiment

* NATS tracing (uses [go-trace](https://github.com/fredbi/go-trace/tree/master/tracer))
* In-deployment, embedded DB migrations (no need for a k8s job to coordinate the deployment) (uses [gooseplus](https://github.com/fredbi/gooseplus))
* Multi-modules configuration, 12-factor app made simple (uses [go-cli/config](https://github.com/fredbi/go-cli/tree/master/config))
* Advanced CLI with cobra (e.g. help topics, dependency injection) (uses [go-cli/cli](https://github.com/fredbi/go-cli/tree/master/cli))
* Streamed DB queries using generic iterators  (uses [go-patterns/iterators](https://github.com/fredbi/go-patterns/tree/master/iterators)
* Playing with decimals (uses cockroachdb's [apd.Decimal](https://github.com/cockroachdb/apd))
* Interesting musings from other people:
  - https://github.com/ripienaar/nats-kv-leader-elect

### Async message processing

* There are N+P _logical_ participants to a message processing facility
* N participants only submit messages to the facility
* P participants only consume messages from the facility
* 1 _logical_ participant may submit or subscribe to the messaging facility from many different computers

* Message processing is idempotent: sending the same message twice doesn't change anything
* Messages have a different subjet, corresponding to the P consumers (P subjects are defined for such routing)
* All messages indicate their submitter and consumer subject (n,p)

* We want submitters to receive feedback that their message has been processed by the appropriate consumer.
* Feedback comes in 3 forms:
    * ACK: the message has been received and is being processed
    * REJECTED: the result of the processing is a rejection, with some cause
    * OK: the result of processing the message is accepted

* No message must be lost
* All messages eventually get a final outcome status

ok, so far this seems to be the perfect use case for a classical pub/sub hub with ACK:

* define 1 topic for POSTINGS, and 1 topic for RESULTS
* define P subjects for POSTINGS
* define N subjects for RESULTS

* if a submitter doesn't get a response within a reasonable time, the message is replayed
  * see how we can minimize what's replayed
  * see how we may use JetStream for replays
* if a submitter doesn't get a processing outcome within a reasonable time, the message is replayed, or eventually cancelled

**NOTE**: consumers and producer *NEVER* give up. Introducing the possibility for backing off at some point
introduces substantial changes to the protocol.

### Variants
* a) the N+P databases are independant
* b) there is one single (possibly sharded) database shared by all participants (considered here as _proxy agents_ to the real participants)

Q: to what extent the setup (b) does simplify integrity?

### Requirements
* participants can query their database _at every moment_ and get an accurate view of the current outcome of their messages
