Marathon Workers
================

## Create CSV From Filters Worker

This worker queries the PUSH_DB using the job filters and builds a CSV file containing user ids that will receive this push notification. Finally, it uploads this CSV file to AWS S3 and calls the next worker (create batches from csv worker).

Only one worker will do this job, but this worker will start `workers.createBatchesFromFilters.pageProcessingConcurrency` goroutines. Each goroutine will get the ids from the database (the metric `get_page_with_filters` will repoirt the retrive duration from the datase and `process_pages` is a counter with the sucessed retrives) and append them to one channel. After this data will be flushed in a CSV file in AWS S3. 

## Create Batches From CSV Worker

This worker downloads a CSV file from AWS S3, reads it and creates batches of user information (locale, token, tz) grouped by timezone. If a job is scheduled and not localized, it schedule all batches in the next worker (process batch worker) for the same timestamp. If a job is scheduled and localized it schedules each batch according to the corresponding timestamp for each timezone. If a job is not schedule it calls the next worker directly for each batch.

Only one worker will do this job, but this worker will start `workers.createBatches.pageProcessingConcurrency` goroutines. Each goroutine will get part of the ids, retrieve the locale and timezone (the metric `get_csv_batch_from_pg` will report the retrieve duration from the database) from the database and schedule each batch (see above).

## Process Batch Worker

This worker receives a batch of user information (locale and token), builds the template for each user using the locale information and the job template name and send to the Kafka topic corresponding to the job app and service. If the error rate is more than a threshold this job enters circuit break state. When the job is paused or in circuit break the batches are stored in a paused job list in Redis with an expiration of one week.

When this work start, it sends thee `starting_process_batch_worker` metric.

This worker send messages to the Kafka. When the message is successful delivery or an error happen, the metric `send_message_return` will be produced.

## Job Completed Worker

When all `Process Batch Worker` is completed, it will call this worker. It will send one email saying the job is completed.

## Resume Job Worker

This worker handles jobs that are paused or in circuit break state. It removes a batch from the paused job list and calls the process batch worker for each one of them until are has no more paused batches.
