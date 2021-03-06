package main

import (
	"fmt"
	"testing"
)

func testHelper(t *testing.T, line []byte, testCase []*testUnit) {
	exporter := newRsyslogExporter()
	exporter.handleStatLine(line)

	for _, k := range exporter.keys() {
		t.Logf("have key: '%s'", k)
	}

	for _, item := range testCase {
		p, err := exporter.get(item.key())
		if err != nil {
			t.Error(err)
		}

		if want, got := item.Val, p.promValue(); want != got {
			t.Errorf("%s: want '%f', got '%f'", item.Name, want, got)
		}
	}

	exporter.handleStatLine(line)

	for _, item := range testCase {
		p, err := exporter.get(item.key())
		if err != nil {
			t.Error(err)
		}

		var wanted float64
		switch p.Type {
		case counter:
			wanted = item.Val
		case gauge:
			wanted = item.Val
		default:
			t.Errorf("%d is not a valid metric type", p.Type)
			continue
		}

		if want, got := wanted, p.promValue(); want != got {
			t.Errorf("%s: want '%f', got '%f'", item.Name, want, got)
		}
	}
}

type testUnit struct {
	Name       string
	Val        float64
	LabelValue string
}

func (t *testUnit) key() string {
	return fmt.Sprintf("%s.%s", t.Name, t.LabelValue)
}

func TestHandleLineWithAction(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "action_processed",
			Val:        100000,
			LabelValue: "test_action",
		},
		&testUnit{
			Name:       "action_failed",
			Val:        2,
			LabelValue: "test_action",
		},
		&testUnit{
			Name:       "action_suspended",
			Val:        1,
			LabelValue: "test_action",
		},
		&testUnit{
			Name:       "action_suspended_duration",
			Val:        1000,
			LabelValue: "test_action",
		},
		&testUnit{
			Name:       "action_resumed",
			Val:        1,
			LabelValue: "test_action",
		},
	}

	actionLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"test_action","processed":100000,"failed":2,"suspended":1,"suspended.duration":1000,"resumed":1}`)
	testHelper(t, actionLog, tests)
}

func TestHandleLineWithResource(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "resource_utime",
			Val:        10,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_stime",
			Val:        20,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_maxrss",
			Val:        30,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_minflt",
			Val:        40,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_majflt",
			Val:        50,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_inblock",
			Val:        60,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_oublock",
			Val:        70,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_nvcsw",
			Val:        80,
			LabelValue: "resource-usage",
		},
		&testUnit{
			Name:       "resource_nivcsw",
			Val:        90,
			LabelValue: "resource-usage",
		},
	}

	resourceLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"resource-usage","utime":10,"stime":20,"maxrss":30,"minflt":40,"majflt":50,"inblock":60,"oublock":70,"nvcsw":80,"nivcsw":90}`)
	testHelper(t, resourceLog, tests)
}

func TestHandleLineWithInput(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "input_submitted",
			Val:        1000,
			LabelValue: "test_input",
		},
	}

	inputLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"test_input", "origin":"imuxsock", "submitted":1000}`)
	testHelper(t, inputLog, tests)
}

func TestHandleLineWithQueue(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "queue_size",
			Val:        10,
			LabelValue: "main Q",
		},
		&testUnit{
			Name:       "queue_enqueued",
			Val:        20,
			LabelValue: "main Q",
		},
		&testUnit{
			Name:       "queue_full",
			Val:        30,
			LabelValue: "main Q",
		},
		&testUnit{
			Name:       "queue_discarded_full",
			Val:        40,
			LabelValue: "main Q",
		},
		&testUnit{
			Name:       "queue_discarded_not_full",
			Val:        50,
			LabelValue: "main Q",
		},
		&testUnit{
			Name:       "queue_max_size",
			Val:        60,
			LabelValue: "main Q",
		},
	}

	queueLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"main Q","size":10,"enqueued":20,"full":30,"discarded.full":40,"discarded.nf":50,"maxqsize":60}`)
	testHelper(t, queueLog, tests)
}

func TestHandleLineWithGlobal(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "dynstat_global",
			Val:        1,
			LabelValue: "msg_per_host.ops_overflow",
		},
		&testUnit{
			Name:       "dynstat_global",
			Val:        3,
			LabelValue: "msg_per_host.new_metric_add",
		},
		&testUnit{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: "msg_per_host.no_metric",
		},
		&testUnit{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: "msg_per_host.metrics_purged",
		},
		&testUnit{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: "msg_per_host.ops_ignored",
		},
	}

	log := []byte(`2018-01-18T09:39:12.763025+00:00 some-node.example.org rsyslogd-pstats: { "name": "global", "origin": "dynstats", "values": { "msg_per_host.ops_overflow": 1, "msg_per_host.new_metric_add": 3, "msg_per_host.no_metric": 0, "msg_per_host.metrics_purged": 0, "msg_per_host.ops_ignored": 0 } }`)

	testHelper(t, log, tests)
}

func TestHandleUnknown(t *testing.T) {
	unknownLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"a":"b"}`)

	exporter := newRsyslogExporter()
	exporter.handleStatLine(unknownLog)

	if want, got := 0, len(exporter.keys()); want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}
}

func TestHandleLineWithImjournal(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "imjournal_submitted",
			Val:        1994,
			LabelValue: "test_imjournal",
		},
		&testUnit{
			Name:       "imjournal_read",
			Val:        1996,
			LabelValue: "test_imjournal",
		},
		&testUnit{
			Name:       "imjournal_discarded",
			Val:        1,
			LabelValue: "test_imjournal",
		},
		&testUnit{
			Name:       "imjournal_failed",
			Val:        1,
			LabelValue: "test_imjournal",
		},
		// &testUnit{
		// 	Name:       "imjournal_poll_failed",
		// 	Val:        0,
		// 	LabelValue: "test_imjournal",
		// },
		&testUnit{
			Name:       "imjournal_rotations",
			Val:        32,
			LabelValue: "test_imjournal",
		},
		&testUnit{
			Name:       "imjournal_recovery_attempts",
			Val:        5,
			LabelValue: "test_imjournal",
		},
		// &testUnit{
		// 	Name:       "imjournal_ratelimit_discarded_in_interval",
		// 	Val:        0,
		// 	LabelValue: "test_imjournal",
		// },
		&testUnit{
			Name:       "imjournal_disk_usage_bytes",
			Val:        75501568,
			LabelValue: "test_imjournal",
		},
	}

	imjournalLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: { "name": "test_imjournal", "origin": "imjournal", "submitted": 1994, "read": 1996, "discarded": 1, "failed": 1, "poll_failed": 0, "rotations": 32, "recovery_attempts": 5, "ratelimit_discarded_in_interval": 0, "disk_usage_bytes": 75501568 }`)
	testHelper(t, imjournalLog, tests)
}

func TestHandleLineWithOmelasticsearch(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "omelasticsearch_submitted",
			Val:        772,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_failedhttp",
			Val:        10,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_failedhttprequests",
			Val:        11,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_failedcheckconn",
			Val:        12,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_failedes",
			Val:        13,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responsesuccess",
			Val:        700,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responsebad",
			Val:        1,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responseduplicate",
			Val:        2,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responsebadargument",
			Val:        3,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responsebulkrejection",
			Val:        4,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_responseother",
			Val:        5,
			LabelValue: "test_omelasticsearch",
		},
		&testUnit{
			Name:       "omelasticsearch_rebinds",
			Val:        6,
			LabelValue: "test_omelasticsearch",
		},
	}

	omelasticsearchLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {
		"name": "test_omelasticsearch", "origin": "omelasticsearch", "submitted": 772, "failed.http": 10,
		"failed.httprequests": 11, "failed.checkConn": 12, "failed.es": 13, "response.success": 700,
		"response.bad": 1, "response.duplicate": 2, "response.badargument": 3, "response.bulkrejection": 4,
		"response.other": 5, "rebinds": 6 }`)
	testHelper(t, omelasticsearchLog, tests)
}

func TestHandleLineWithMmkubernetes(t *testing.T) {
	tests := []*testUnit{
		&testUnit{
			Name:       "mmkubernetes_recordseen",
			Val:        9876,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacemetadatasuccess",
			Val:        11,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacemetadatanotfound",
			Val:        1,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacemetadatabusy",
			Val:        2,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacemetadataerror",
			Val:        3,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podmetadatasuccess",
			Val:        12,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podmetadatanotfound",
			Val:        4,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podmetadatabusy",
			Val:        5,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podmetadataerror",
			Val:        6,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacecachenumentries",
			Val:        13,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podcachenumentries",
			Val:        14,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacecachehits",
			Val:        15,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podcachehits",
			Val:        16,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_namespacecachemisses",
			Val:        17,
			LabelValue: "test_mmkubernetes",
		},
		&testUnit{
			Name:       "mmkubernetes_podcachemisses",
			Val:        18,
			LabelValue: "test_mmkubernetes",
		},
	}

	mmkubernetesLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {
		"name": "test_mmkubernetes",
		"origin": "mmkubernetes", "recordseen": 9876, "namespacemetadatasuccess": 11, "namespacemetadatanotfound": 1,
		"namespacemetadatabusy": 2, "namespacemetadataerror": 3, "podmetadatasuccess": 12, "podmetadatanotfound": 4,
		"podmetadatabusy": 5, "podmetadataerror": 6, "namespacecachenumentries": 13, "podcachenumentries": 14,
		"namespacecachehits": 15, "podcachehits": 16, "namespacecachemisses": 17, "podcachemisses": 18 }`)
	testHelper(t, mmkubernetesLog, tests)
}
