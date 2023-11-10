# nats

I want to experiment with NATS, and deploy a transactional round-trip message processing pattern.

Here is the storyboard:

(n) ---> (p) ----
                |
(feedback)   <---

## Async message processing

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

(n)  -> (publish to POSTINGS) - |
                                | -        (subscribe with subject) (p)
                                                      |
                                               (consume from topic)
                                                  ACK (consumed)
                                                      |
                                         (publish ACK to RESULTS with subject n)
                                                    (process N)
                                                      |
                                         (publish outcome to RESULTS with subject n)
                                                      |
(n)  -> (subscribe to RESULTS with subject) -         |
            ACK (consumed)
             (process P)

ok so the pipeline is fully async.

* if a submitter doesn't get a ACK within a reasonable time, the message is replayed, or eventually cancelled after n attempts
  * see how we can minimize what's replayed
  * see how we may use JetStream for replays
* if a submitter doesn't get a processing outcome within a reasonable time, the message is replayed, or eventually cancelled


## Interactions with each participant's local state

Each participant to the facility maintains their own state of messages sent and their outcome.
In other words, each participant has its own database to store the result of their activity.

### Variants
* a) the N+P databases are independant
* b) there is one single (possibly sharded) database shared by all participants (considered here as _proxy agents_ to the real participants)

Q: to what extent the setup (b) does simplify integrity?

### Requirements
* participants can query their database _at every moment_ and get an accurate view of the current outcome of their messages

## Edge cases
1. lost ACK
* message X has been posted by n to p
* n acked the message, but the ACK message is never received by n

=> after t1, (n) replays the message, received again by n, who discards it if it has already been processed

1. lost response
* message X has been posted by n to p
* n processed the message, but the response is lost

=> after t2, (n) replays the message, received again by n, who discards it if it has already been processed

2. lost processing
* message X has been posted by n to p
* n acked the message, but processing failed without response

=> after t2, (n) replays the message, received again by n, who attempts again to process it

3. failed local state updates

* message X has been posted by n to p
* status is sent back to n, who fails to update its state

=> local DB tx is not committed, and the subscribed message is not acked. When (n) is up again, the message is delivered

* message X has been posted by n to p
* (p) fails to update its state

=> the message should not be ACK'ed from the subscription before (p) has updated its state

4. Interesting variants

* messages as  arrays  of unitary messages
* can the extra assumption that message is identified by a monotonous sequence be used to simplify the acknoledgement protocol?

5. Participants are "agents"

If we assume that our facility is ony made of proxy agents, which are owned by a single authority,
then each proxy has to forward the message, possibly using a different protocol (file transfer, etc) to the actual participant.

Protocol breaks are put under the "processing" abstraction.
