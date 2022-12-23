package worker

const (
	createBatchesWorkerStart     = "starting_create_batches_worker"
	createBatchesWorkerCompleted = "completed_create_batches_worker"
	createBatchesWorkerError     = "error_create_batches_worker"

	csvSplitWorkerStart     = "starting_csv_split_worker"
	csvSplitWorkerCompleted = "completed_csv_split_worker"
	csvSplitWorkerError     = "error_csv_split_worker"

	directWorkerStart     = "starting_direct_part"
	directWorkerCompleted = "completed_direct_worker"
	directWorkerError     = "error_direct_worker"

	jobCompletedWorkerStart     = "starting_job_completed_worker"
	jobCompletedWorkerCompleted = "completed_job_completed_worker"
	jobCompletedWorkerError     = "error_job_completed_worker"

	processBatchWorkerStart     = "starting_process_batch_worker"
	processBatchWorkerCompleted = "completed_process_batch_worker"
	processBatchWorkerError     = "error_process_batch_worker"

	resumeJobWorkerStart     = "starting_resume_job_worker"
	resumeJobWorkerCompleted = "completed_resume_job_worker"
	resumeJobWorkerError     = "error_resume_job_worker"

	getCsvFromS3Timing = "get_csv_from_s3"
)
