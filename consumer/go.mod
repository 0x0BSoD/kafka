module github.com/0x0bsod/kafka/consumer

go 1.15

require (
	github.com/0x0bsod/goLogz v0.0.0-20200804084250-e4ff4744945c
	github.com/0x0bsod/kafka/common v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.9.0
	github.com/segmentio/kafka-go v0.4.9
)

replace github.com/0x0bsod/kafka/common => ../common
