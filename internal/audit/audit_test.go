package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TFK70/kube-apiserver-audit-exporter/internal/logging"
)

var exampleLines = []string{
	`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"bf4e6c0a-ee95-48e9-b86b-93009c13f7d7","stage":"ResponseComplete","requestURI":"/apis/coordination.k8s.io/v1/namespaces/monitoring/leases/f75f3bba.integreatly.org?timeout=5s","verb":"update","user":{"username":"system:serviceaccount:monitoring:grafana-operator","uid":"722e6d5c-bf78-45b5-b3f2-0e7a74993c49","groups":["system:serviceaccounts","system:serviceaccounts:monitoring","system:authenticated"],"extra":{}},"sourceIPs":["192.168.100.100"],"userAgent":"v5/v0.0.0 (linux/amd64) kubernetes/$Format/leader-election","objectRef":{"resource":"leases","namespace":"ns","name":"f75f3bba.integreatly.org","uid":"a868bdb5-0fa5-41a7-b468-da0a1a46876b","apiGroup":"coordination.k8s.io","apiVersion":"v1","resourceVersion":"4363907"},"responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2025-07-18T20:39:27.753542Z","stageTimestamp":"2025-07-18T20:39:27.760310Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"grafana-operator\" of ClusterRole \"grafana-operator\" to ServiceAccount \"grafana-operator/monitoring\""}}`,
	`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"bf4e6c0a-ee95-48e9-b86b-93009c13f7d7","stage":"RequestReceived","requestURI":"/apis/coordination.k8s.io/v1/namespaces/monitoring/leases/f75f3bba.integreatly.org?timeout=5s","verb":"update","user":{"username":"system:serviceaccount:monitoring:grafana-operator","uid":"722e6d5c-bf78-45b5-b3f2-0e7a74993c49","groups":["system:serviceaccounts","system:serviceaccounts:monitoring","system:authenticated"],"extra":{}},"sourceIPs":["192.168.100.100"],"userAgent":"v5/v0.0.0 (linux/amd64) kubernetes/$Format/leader-election","objectRef":{"resource":"leases","namespace":"ns","name":"f75f3bba.integreatly.org","apiGroup":"coordination.k8s.io","apiVersion":"v1"},"requestReceivedTimestamp":"2025-07-18T20:39:27.753542Z","stageTimestamp":"2025-07-18T20:39:27.753542Z"}`,
	`{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"5191fd8f-4881-44f4-807f-bb3005589a40","stage":"ResponseComplete","requestURI":"/livez","verb":"get","user":{"username":"system:serviceaccount:monitoring:vmk8s-kube-state-metrics","uid":"dd0e0787-96a8-4b36-b112-58f00d8061c4","groups":["system:serviceaccounts","system:serviceaccounts:monitoring","system:authenticated"],"extra":{}},"sourceIPs":["192.168.100.100"],"userAgent":"kube-state-metrics/v2.15.0 (linux/amd64) kubernetes/","responseStatus":{"metadata":{},"code":200},"requestReceivedTimestamp":"2025-07-18T20:39:27.750543Z","stageTimestamp":"2025-07-18T20:39:27.752798Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":"RBAC: allowed by ClusterRoleBinding \"system:discovery\" of ClusterRole \"system:discovery\" to Group \"system:authenticated\""}}`,
}

var mu *sync.Mutex = &sync.Mutex{}

const (
	TEST_DIR = "testdata"
	TEST_FILE = "testdata/audit.log"
)

func cleanup() error {
	err := os.Truncate(TEST_FILE, 0)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %v", err)
	}

	return nil
}

func writeLines(t *testing.T) chan struct{} {
	done := make(chan struct{})

	file, err := os.OpenFile(TEST_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	go func() {
		for _, line := range exampleLines {
			if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
				t.Logf("failed to write line: %v", err)
			}

			time.Sleep(time.Millisecond * 100)
		}

		done <- struct{}{}
		close(done)
		file.Close()
	}()

	return done
}

func TestAuditReader(t *testing.T) {
	mu.Lock()
	defer cleanup()
	defer mu.Unlock()

	logging.SetupLogger()
	logging.NullifyLogger()

	reader, err := NewReader(
		WithPath(TEST_FILE),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = reader.Start()
	if err != nil {
		t.Fatal(err)
	}

	expectedLines := []Event{}

	for _, line := range exampleLines {
		var event Event
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			t.Fatal(err)
		}

		expectedLines = append(expectedLines, event)
	}

	receivedLines := []Event{}

	doneReading := make(chan struct{})
	go func() {
		for log := range reader.Events {
			receivedLines = append(receivedLines, log)
		}

		doneReading <- struct{}{}
	}()

	doneWriting := writeLines(t)

	<-doneWriting

	reader.Stop()

	<-doneReading

	assert.Equal(t, len(expectedLines), len(receivedLines))
	assert.ElementsMatch(t, expectedLines, receivedLines)
}


func TestAuditReaderWithRename(t *testing.T) {
	mu.Lock()
	defer cleanup()
	defer mu.Unlock()

	logging.SetupLogger()
	logging.NullifyLogger()

	reader, err := NewReader(
		WithPath(TEST_FILE),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = reader.Start()
	if err != nil {
		t.Fatal(err)
	}

	expectedLines := []Event{}

	for range [2]int{} {
		for _, line := range exampleLines {
			var event Event
			err := json.Unmarshal([]byte(line), &event)
			if err != nil {
				t.Fatal(err)
			}

			expectedLines = append(expectedLines, event)
		}
	}

	receivedLines := []Event{}

	doneReading := make(chan struct{})
	go func() {
		for log := range reader.Events {
			receivedLines = append(receivedLines, log)
		}

		doneReading <- struct{}{}
	}()

	doneWriting1 := writeLines(t)

	<-doneWriting1

	err = os.Rename(TEST_FILE, TEST_FILE + ".renamed")
	if err != nil {
		t.Fatalf("failed to rename file: %v", err)
	}

	file, err := os.Create(TEST_FILE)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	file.Close()

	time.Sleep(time.Second * 2) // wait for reader to re-attach to file

	doneWriting2 := writeLines(t)

	<-doneWriting2

	reader.Stop()

	<-doneReading

	assert.Equal(t, len(expectedLines), len(receivedLines))
	assert.ElementsMatch(t, expectedLines, receivedLines)

	err = os.Remove(TEST_FILE + ".renamed")
	if err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}
}
