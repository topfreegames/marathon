package worker

const (
	CreateBatchesWorkerStart     = "starting_create_batches_worker"
	CreateBatchesWorkerCompleted = "completed_create_batches_worker"
	CreateBatchesWorkerError     = "error_create_batches_worker"

	CsvSplitWorkerStart     = "starting_csv_split_worker"
	CsvSplitWorkerCompleted = "completed_csv_split_worker"
	CsvSplitWorkerError     = "error_csv_split_worker"

	DirectWorkerStart     = "starting_direct_part"
	DirectWorkerCompleted = "completed_direct_worker"
	DirectWorkerError     = "error_direct_worker"

	JobCompletedWorkerStart     = "starting_job_completed_worker"
	JobCompletedWorkerCompleted = "completed_job_completed_worker"
	JobCompletedWorkerError     = "error_job_completed_worker"

	ProcessBatchWorkerStart     = "starting_process_batch_worker"
	ProcessBatchWorkerCompleted = "completed_process_batch_worker"
	ProcessBatchWorkerError     = "error_process_batch_worker"

	ResumeJobWorkerStart     = "starting_resume_job_worker"
	ResumeJobWorkerCompleted = "completed_resume_job_worker"
	ResumeJobWorkerError     = "error_resume_job_worker"

	GetCsvFromS3Timing   = "get_csv_from_s3"
	GetUsersFromDbTiming = "get_from_pg"
)
