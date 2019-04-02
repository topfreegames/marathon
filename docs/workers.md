Marathon Workers
================

## Create CSV From Filters Worker

This worker queries the PUSH_DB using the job filters and creates the tasks for the next worker.  It will retrieve information from the database to create the necessary queries to produce the batchs in the next worker.

The first step of this worker is to count how many elements it needs to process. The running time of this query is present in the metric `count_total`.

The metric `get_intervals_cursor` is the total time spend to collect the batchs intervals.
 
## DB to CSV Worker

This worker will get the batches queries from the previous worker and get the information from the database. It will also upload the list of  `users_id`  to the Amazon, using the multipart upload. After each part upload, the result will be saved in a Redis list (the name of the list is the job UUID). When all parts are successful uploaded, the last worker will send the complete multipart upload event to Amazon and create the next task.

The metric `get_page_with_filters` will measure the time spent to retrieve each batch from the database. And the `write_user_page_into_csv`  count how many times one S3 part upload started.

## Create Batches From CSV Worker

This worker downloads a CSV file from AWS S3, reads it and creates batches of user information (locale, token, tz) grouped by timezone. If a job is scheduled and not localized, it schedules all batches in the next worker (process batch worker) for the same timestamp. If a job is scheduled and localized it schedules each batch according to the corresponding timestamp for each timezone. If a job is not scheduled, it calls the next worker directly for each batch.

Only one worker will do this job, but this worker will start `workers.createBatches.pageProcessingConcurrency` goroutines. Each goroutine will get part of the ids, retrieve the locale and timezone (the metric `get_csv_batch_from_pg` will report the retrieve duration from the database) from the database and schedule each batch (see above).

## Process Batch Worker

This worker receives a batch of user information (locale and token), builds the template for each user using the locale information and the job template name and send to Kafka topic corresponding to the job app and service. If the error rate is more than a threshold this job enters circuit break state. When the job is paused or in circuit break the batches are stored in a paused job list in Redis with an expiration of one week.

When this work start, it sends the `starting_process_batch_worker` metric.

This worker sends messages to the Kafka. When the message is delivered, either successfully or with errors, the metric `send_message_return` will be produced.

## Job Completed Worker

When all `Process Batch Worker` is completed, it will call this worker. It will send one email saying the job is completed.

## Resume Job Worker

This worker handles jobs that are paused or in circuit break state. It removes a batch from the paused job list and calls the process batch worker for each one of them until are has no more paused batches.
