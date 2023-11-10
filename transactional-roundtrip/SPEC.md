# specs

## Pub/sub infrastructure

* topic POSTINGS
* topic RESULTS

* producers publish to POSTINGS and subscribe to RESULTS
* consumers subscribe to POSTINGS and publish to RESULTS

## Messages

 |   Field    |  Description |
 |------------|--------------|
 | header     |  n, p identifiers |
 | message ID |  must be unique for every sender n, does not need to be globally unique |
 | status     |  the current status of the message, as a pair (ack status, processing status), as known from the last updater |
 | replays    |  how many times this message has been replayed |
 | inception_datum |  ts of first posting |
 | last_datum |  ts of last replay |
 | payload    |  arbitrary |
 | audit      |  audit trail of various actions & processing |

### Message statuses
ACK status:
* nacked posting, nacked result
* acked posting, nacked posting

Processing status:
* PENDING
* REJECTED
* OK

## Databases

every participant maintains a Postgres DB (variant cockroach DB cluster).
every participant runs x independant instances against this DB

## Protocol

1. Producer (n) posting
   * upon startup: look in DB for messages in a pending status, replay them
     * for unacked messages, repeat the sequence below (with DB update)
     * for acked messages, call the facility for message replay, update status in DB

   * can do the following in parallel for x messages:

   * populate a new message M with unique ID, for correspondant (n)
   * DB TX{record M with # retries & updated ts}                                           <---|
     [status is (nacked posting, pending)]

CRITICAL SECTION
   * publish M to POSTINGS with subject (p) ; wait for ACK -> retry until ACK or timeout/too many attempts  ----|
   * if bailed from publishing {
        updated DB with actual # retries & ts
     }
   * update DB with status (acked, processing status existing in DB)

   * subscribe to RESULTS on subject (n):
      * receive message
      * update (acked result) in DB: if not found RAISE CONSISTENCY ISSUE
      * ack message from subscription

2. Consumer (p) processing:
   * upon startup

   * subscribe to POSTINGS with subject (p)

CRITICAL SECTION
