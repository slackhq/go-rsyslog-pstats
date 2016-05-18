package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_parseMsg(t *testing.T) {
	var lines []string
	res := bytes.NewBufferString("")

	// Ignore unparseable lines
	res.Reset()
	parseMsg([]byte("May 16 20:51:58 hostname rsyslogd-pstats: @ { asdf"), res)
	assert.Equal(t, 0, res.Len())

	// Ignore empty dynstats
	res.Reset()
	parseMsg([]byte("May 16 20:51:58 hostname rsyslogd-pstats: @cee: { \"name\": \"global\", \"origin\": \"dynstats\", \"values\": { } }"), res)
	assert.Equal(t, 0, res.Len())

	// Parse and sanitize dynstats properly
	res.Reset()
	parseMsg([]byte("May 16 20:51:58 hostname rsyslogd-pstats: @cee: { \"name\": \"global\", \"origin\": \"dynstats\", \"values\": { \"thing.one\": 100  }}"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 1)
	assert.Contains(t, lines, "rsyslog.dynstats.thing_one:100|g")

	// Make sure we remove extraneous dots
	res.Reset()
	parseMsg([]byte("May 16 20:51:58 hostname rsyslogd-pstats: {\"name\":\"imuxsock\",\"origin\":\"imuxsock\",\"submitted\":8,\"ratelimit.discarded\":0,\"ratelimit.numratelimiters\":0}"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, lines, "rsyslog.imuxsock.imuxsock.submitted:8|g")
	assert.Contains(t, lines, "rsyslog.imuxsock.imuxsock.ratelimit_discarded:0|g")
	assert.Contains(t, lines, "rsyslog.imuxsock.imuxsock.ratelimit_numratelimiters:0|g")

	// Make sure we strip out crazy garbage that makes the stat name worse
	res.Reset()
	parseMsg([]byte("May 16 20:51:58 hostname rsyslo { \"name\": \"weird_input(*//var/run/sock/IPv4)\", \"origin\": \"imptcp\", \"submitted\": 1000234982, \"bytes.received\": 0, \"bytes.decompressed\": 0 }"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, lines, "rsyslog.imptcp.weird_input_var_run_sock_ipv4.submitted:1000234982|g")
	assert.Contains(t, lines, "rsyslog.imptcp.weird_input_var_run_sock_ipv4.bytes_received:0|g")
	assert.Contains(t, lines, "rsyslog.imptcp.weird_input_var_run_sock_ipv4.bytes_decompressed:0|g")

	// For good measure
	res.Reset()
	parseMsg([]byte("{\"name\":\"resource-usage\",\"origin\":\"impstats\",\"utime\":0,\"stime\":8000,\"maxrss\":4676,\"minflt\":420,\"majflt\":9,\"inblock\":152,\"oublock\":40,\"nvcsw\":44,\"nivcsw\":6}"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 9)
	assert.Contains(t, lines, "rsyslog.resource_usage.nivcsw:6|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.utime:0|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.maxrss:4676|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.minflt:420|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.majflt:9|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.stime:8000|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.inblock:152|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.oublock:40|g")
	assert.Contains(t, lines, "rsyslog.resource_usage.nvcsw:44|g")

	// Make sure we lower case everything
	res.Reset()
	parseMsg([]byte("e: {\"name\":\"main Q\",\"origin\":\"core.queue\",\"size\":13,\"enqueued\":95,\"full\":0,\"discarded.full\":0,\"discarded.nf\":0,\"maxqsize\":14}"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 6)
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.discarded_nf:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.maxqsize:14|g")
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.size:13|g")
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.enqueued:95|g")
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.full:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.main_q.discarded_full:0|g")

	// Make sure we remove crazy characters and remove any underscores that appear at the end of the line
	res.Reset()
	parseMsg([]byte("{ \"name\": \"action 1 queue[DA]\", \"origin\": \"core.queue\", \"size\": 0, \"enqueued\": 0, \"full\": 0, \"discarded.full\": 0, \"discarded.nf\": 0, \"maxqsize\": 0 }"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 6)
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.size:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.enqueued:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.full:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.discarded_full:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.discarded_nf:0|g")
	assert.Contains(t, lines, "rsyslog.core_queue.action_1_queue_da.maxqsize:0|g")

	// Make sure actions are handled properly
	res.Reset()
	parseMsg([]byte("{ \"name\": \"custom_relp\", \"origin\": \"core.action\", \"processed\": 1223, \"failed\": 0, \"suspended\": 0, \"suspended.duration\": 0, \"resumed\": 0 }"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 5)
	assert.Contains(t, lines, "rsyslog.core_action.custom_relp.failed:0|g")
	assert.Contains(t, lines, "rsyslog.core_action.custom_relp.suspended:0|g")
	assert.Contains(t, lines, "rsyslog.core_action.custom_relp.suspended_duration:0|g")
	assert.Contains(t, lines, "rsyslog.core_action.custom_relp.resumed:0|g")
	assert.Contains(t, lines, "rsyslog.core_action.custom_relp.processed:1223|g")

	// Make sure weird input names are handled properly
	res.Reset()
	parseMsg([]byte("{ \"name\": \"imrelp[0333]\", \"origin\": \"imrelp\", \"submitted\": 7092077 }"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 1)
	assert.Contains(t, lines, "rsyslog.imrelp.imrelp_0333.submitted:7092077|g")

	// Make sure io-work-q is handled properly
	res.Reset()
	parseMsg([]byte("{ \"name\": \"io-work-q\", \"origin\": \"imptcp\", \"enqueued\": 0, \"maxqsize\": 0 }"), res)
	lines = strings.Split(strings.Trim(res.String(), "\n"), "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines, "rsyslog.imptcp.io_work_q.enqueued:0|g")
	assert.Contains(t, lines, "rsyslog.imptcp.io_work_q.maxqsize:0|g")
}

func Benchmark_parseMsg(b *testing.B) {
	res := bytes.NewBufferString("")
	msg := []byte("{ \"name\": \"action 1 queue[DA]\", \"origin\": \"core.queue\", \"size\": 0, \"enqueued\": 0, \"full\": 0, \"discarded.full\": 0, \"discarded.nf\": 0, \"maxqsize\": 0 }")

	for i := 0; i < b.N; i++ {
		parseMsg(msg, res)
		res.Reset()
	}
}
