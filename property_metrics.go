package main

import (
	"fmt"
	"time"
)

type MaintenanceRequest struct {
	ID        string
	Open      bool
	Due       time.Time
	Completed time.Time
}

type TenantDocument struct {
	TenantID string
	Current  bool
}

type InspectionReminder struct {
	UnitID string
	Due    time.Time
}

type PropertySnapshot struct {
	Requests    []MaintenanceRequest
	Documents   []TenantDocument
	Inspections []InspectionReminder
	Now         time.Time
}

type BusinessMetrics struct {
	OpenMaintenance  int
	MissingDocuments int
	DueInspections   int
}

func ComputeMetrics(snapshot PropertySnapshot) BusinessMetrics {
	result := BusinessMetrics{}
	for _, request := range snapshot.Requests {
		if request.Open && request.Completed.IsZero() {
			result.OpenMaintenance++
		}
	}
	for _, document := range snapshot.Documents {
		if !document.Current {
			result.MissingDocuments++
		}
	}
	for _, inspection := range snapshot.Inspections {
		if !inspection.Due.After(snapshot.Now) {
			result.DueInspections++
		}
	}
	return result
}

func PublishMetrics(client *MetricsClient, snapshot PropertySnapshot) (BusinessMetrics, error) {
	metrics := ComputeMetrics(snapshot)
	tags := map[string]string{"service": "property-management"}
	reports := []struct {
		name  string
		value int
	}{
		{"property.maintenance.open", metrics.OpenMaintenance},
		{"property.documents.missing", metrics.MissingDocuments},
		{"property.inspections.due", metrics.DueInspections},
	}
	for _, report := range reports {
		if err := client.Report(report.name, report.value, tags, fmt.Sprintf("property-metrics-%s", snapshot.Now.UTC().Format("20060102150405"))); err != nil {
			return metrics, err
		}
	}
	return metrics, nil
}

func main() {
	client, err := NewMetricsClient()
	if err != nil {
		panic(err)
	}
	now := time.Now().UTC()
	snapshot := PropertySnapshot{
		Now:         now,
		Requests:    []MaintenanceRequest{{ID: "req-17", Open: true}, {ID: "req-18", Open: false}},
		Documents:   []TenantDocument{{TenantID: "tenant-4", Current: false}, {TenantID: "tenant-9", Current: true}},
		Inspections: []InspectionReminder{{UnitID: "unit-2", Due: now.Add(-time.Hour)}},
	}
	metrics, err := PublishMetrics(client, snapshot)
	if err != nil {
		panic(err)
	}
	fmt.Printf("reported open_maintenance=%d missing_documents=%d due_inspections=%d\n", metrics.OpenMaintenance, metrics.MissingDocuments, metrics.DueInspections)
}
