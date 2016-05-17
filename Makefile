test:
	go test -v

test-cov-html:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out

bench:
	go test -bench=.

bench-cpu:
	go test -bench=. -benchtime=5s -cpuprofile=cpu.pprof
	go tool pprof go-rsyslog-pstats.test cpu.pprof

bench-cpu-long:
	go test -bench=. -benchtime=60s -cpuprofile=cpu.pprof
	go tool pprof go-rsyslog-pstats.test cpu.pprof

deb:
	go build
	sh deb.sh

.PHONY: test test-cov-html bench bench-cpu bench-cpu-long deb
