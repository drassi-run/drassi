package transport

import (
	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/reporter/service"
)

type Factory interface {
	StepLogsCourier(sr executor.StepRun) Courier
	JobLogsCourier() Courier
	DiagnosticLogsCourier() Courier
	AttachmentCourier(kind, name string) Courier
}

type resultFactory struct {
	svc *service.ResultService
}

func (fac *resultFactory) StepLogsCourier(sr executor.StepRun) Courier {
	u := fac.svc.StepLogsUploader(sr)
	return &chunkCourier{uploader: u}
}

func (fac *resultFactory) JobLogsCourier() Courier {
	u := fac.svc.JobLogsUploader()
	return &fileCourier{uploader: u}
}

func (fac *resultFactory) DiagnosticLogsCourier() Courier {
	u := fac.svc.DiagnosticLogsUploader()
	return &fileCourier{uploader: u}
}

func (fac *resultFactory) AttachmentCourier(kind, name string) Courier {
	//if kind == "StepSummary" {
	//	return &fileCourier{uploader: fac.svc.StepSummaryUploader(name)}
	//}
	// TODO
	return nil
}

type jobFactory struct {
	svc *service.JobService
}

func (fac *jobFactory) StepLogsCourier(sr executor.StepRun) Courier {
	recId := ""
	u := fac.svc.LogUploader(recId)
	return &fileCourier{uploader: u}
}

func (fac *jobFactory) JobLogsCourier() Courier {
	recId := ""
	u := fac.svc.LogUploader(recId)
	return &fileCourier{uploader: u}
}

func (fac *jobFactory) DiagnosticLogsCourier() Courier {
	recId := ""
	u := fac.svc.LogUploader(recId)
	return &fileCourier{uploader: u}
}

func (fac *jobFactory) AttachmentCourier(kind, name string) Courier {
	recId := ""
	u := fac.svc.AttachmentUploader(recId, kind, name)
	return &fileCourier{uploader: u}
}
