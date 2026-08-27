package service

import (
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

func TestJobToResponseAggregatesPersistedResults(t *testing.T) {
	job := &model.ScanJob{ID: 9, Status: "completed", Results: `[{"company_name":"Acme","new_roles":2,"total_roles":3},{"company_name":"Broken","error":"offline"}]`}
	response := jobToResponse(job)
	if response.TotalNewRoles != 2 || response.TotalRoles != 3 || len(response.Results) != 2 {
		t.Fatalf("response = %#v", response)
	}
	job.Results = "not-json"
	if got := jobToResponse(job); len(got.Results) != 0 || got.TotalRoles != 0 {
		t.Fatalf("invalid results response = %#v", got)
	}
}

func TestScannerServiceCreatesAndReadsJobs(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	svc := NewScannerService(db, repos)
	id, err := svc.StartScan()
	if err != nil || id == 0 {
		t.Fatalf("start scan = %d, %v", id, err)
	}
	job, err := svc.GetJob(id)
	if err != nil || job == nil {
		t.Fatalf("get created job = %#v, %v", job, err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for job.Status != "completed" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		job, err = svc.GetJob(id)
		if err != nil {
			t.Fatalf("poll job: %v", err)
		}
	}
	if job.Status != "completed" {
		t.Fatalf("scan did not complete: %#v", job)
	}
}
