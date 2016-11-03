Marathon Architecture
=====================

## Introduction

Marathon is divided in three major parts:

* UI for starting new jobs (Philipides);
* Ingestion API;
* Stream Worker.

## Ingestion API

The ingestion API is a multi-tenant service that divides jobs between apps and ensures that only the users with access to those apps can see those jobs.

All routes (apart from creating an app) **must** be scoped per app. The API must be multi-tenant.

It is responsible for:

* Creating a new app;
* Granting access to an app for a specific user;
* Submitting templates;
* Accepting new jobs to be processed (with a CSV of users or a segment of users);
* Storing a job in the Database to be updated by the workers;
* Storing the job as batches in a queue (kue);
* Reporting on the status of a job;
* Stopping a job;
* Allow apps to report user tokens;   // Marathon V3
* Allow one user push notifications.  // Marathon V3

The API should create both an entry in a table for Jobs in the relational Database as well as creating a job in Kue:

### PostgreSQL

```
Jobs Table
id                                      batches     completed_at
8C3090B0-A5F9-4685-B4BB-F04BA001518F    123         2016-10-10 10:10:10

FailedJobs Table
id                                      batch   notification_index      error
8C3090B0-A5F9-4685-B4BB-F04BA001518F    2       3                       Could not connect to apple service: ...
```

Each batch will be stored as a job with the following data:

```
{
    "jobId": "8C3090B0-A5F9-4685-B4BB-F04BA001518F",
    "batch": 2,
    "users": [
        "666EC3D0-6002-4B3B-943E-4BD7E8118E55",
        "45E0AF51-FF32-4F38-B861-D03C0D300C34",
        "1C4A7441-8D4E-446E-9CCC-5A40D274D0CC",
        "D1865F82-BC36-4597-9750-F58D97E6F8D0",
        ...
    ],
    "template": {
        "key": "some-template",     //template key
        "contents": "<%=x>x<%=y>"   //or contents, never both
    },
    "context": {                    // context for interpolation in the template
        "x": 1,
        "y": 2
    },
    "app": "com.fungames.epiccardgamebeta",
    "service": "apns",
    "expiration": 4928148128381284, //Unix Timestamp
}
```

## Stream Worker

The stream worker is a Kue worker responsible for taking the raw batches that were created by the API and transforming them in single push notification inputs for the Notification Worker.

Let's say our batches contain 1000 notifications each (1000 user ids), then the Stream worker will pick up a batch and for each of those 1000 user ids, it will:

* Retrieve the token for all users in the batch;
* Merge the context sent with the batch with the defaults for the given template;
* Compile the template with the context;
* Send the batch as a whole (batch send in kafka lib) to Kafka for processing by other systems;
* Incrementing either successful or failed counters in redis as well as storing failures in the database.
