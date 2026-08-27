package main

import (
	"testing"
	"time"
)

func TestComputeMetricsCountsOperationalWork(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	got := ComputeMetrics(PropertySnapshot{
		Now:         now,
		Requests:    []MaintenanceRequest{{Open: true}, {Open: true, Completed: now}, {Open: false}},
		Documents:   []TenantDocument{{Current: false}, {Current: true}},
		Inspections: []InspectionReminder{{Due: now}, {Due: now.Add(time.Minute)}},
	})
	want := BusinessMetrics{OpenMaintenance: 1, MissingDocuments: 1, DueInspections: 1}
	if got != want {
		t.Fatalf("metrics = %#v, want %#v", got, want)
	}
}
