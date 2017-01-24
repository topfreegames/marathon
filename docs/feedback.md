Marathon Feedbacks
==================

After a message is sent to the kafka topic corresponding to the job app and service, another system will attempt to send the push notification to APNS or GCM and write a feedback of this successful or failed push notification in another queue, the feedback kafka.

Marathon has another command, besides the API and the workers, that starts a feedback listener that reads from this kafka's topics and update the job's feedback column in the PostgreSQL database. The messages in this queue contain all metadata sent by marathon including the job id.

## Feedbacks column

The feedbacks column contains a JSON in the following format:  

```json
{
  "ack":        <int>,  // count
  "error-key1": <int>,  // count
  "error-key2": <int>   // count
  ...
}
```

In the case of successful push notifications the key is `ack`. For failed push notifications the key will be the error reason received from APNS or GCM, for example `BAD_REGISTRATION`, `unregistered`, etc.

To avoid updating the job entry in the PostgreSQL database for every message received in the feedbacks kafka, we update the database periodically (defaults to every 5 seconds) by using a local cache to store all feedbacks received in the mean time.
