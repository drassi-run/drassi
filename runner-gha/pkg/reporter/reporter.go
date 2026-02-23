package reporter

import (
	"context"
	"fmt"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/gha-runner/pkg/service"
	"drassi.run/gha-runner/pkg/types"
	"github.com/google/uuid"
)

func New(recorder service.TimelineRecorder) *GhaReporter {
	return &GhaReporter{
		recorder: recorder,
	}
}

type GhaReporter struct {
	order     int
	jobRecord *types.Record
	records   []*types.Record

	recorder service.TimelineRecorder
}

func (r *GhaReporter) JobRecord() *types.Record {
	return r.jobRecord
}

func (r *GhaReporter) StartJob(ctx context.Context, je executor.JobExecutor) error {
	r.order++

	r.jobRecord = &types.Record{
		Uid:   executor.JobUid(je),
		Order: r.order,
		Object: &types.JobObject{
			JobRun: je.JobRun(),
		},
	}

	now := time.Now()
	r.jobRecord.StartedAt = &now
	r.records = append(r.records, r.jobRecord)

	return r.recorder.Update(ctx, r.jobRecord)
}

func (r *GhaReporter) EndJob(ctx context.Context, je executor.JobExecutor, result *records.Job) error {
	now := time.Now()
	r.jobRecord.CompletedAt = &now
	r.jobRecord.State = types.StateCompleted
	r.jobRecord.Result = types.ToResult(result.Result)

	return r.recorder.Update(ctx, r.jobRecord)
}

func (r *GhaReporter) StartStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor) error {
	if len(r.records) == 0 {
		return fmt.Errorf("no records found for stage %s", stage)
	}

	r.order++

	record := &types.Record{
		Order: r.order,
		Object: &types.StepObject{
			StepRun: se.StepRun(),
			Stage:   stage,
		},
	}
	now := time.Now()
	record.StartedAt = &now

	if stage == executor.StageMain {
		record.Uid = executor.StepUid(se)
	} else {
		record.Uid = uuid.New().String()
	}

	parent := r.records[len(r.records)-1]
	parent.Children = append(parent.Children, record)
	r.records = append(r.records, record)

	if se.ParentExecutor() == nil {
		return r.recorder.Update(ctx, record)
	}
	return nil
}

func (r *GhaReporter) EndStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor, result *records.Step) error {
	if len(r.records) == 0 {
		return fmt.Errorf("no records found for stage %s", stage)
	}

	record := r.records[len(r.records)-1]
	r.records = r.records[:len(r.records)-1]

	now := time.Now()
	record.CompletedAt = &now
	record.State = types.StateCompleted
	record.Result = types.ToResult(result.Conclusion)

	if se.ParentExecutor() == nil {
		return r.recorder.Update(ctx, record)
	}
	return nil
}

func (r *GhaReporter) Close() error {
	//TODO implement me
	panic("implement me")
}
