module github.com/0x0bsod/kafka/proucer

go 1.15

replace github.com/0x0bsod/kafka/common => ../common

require (
	github.com/0x0bsod/goLogz v0.0.0-20200804084250-e4ff4744945c
	github.com/0x0bsod/kafka/common v0.0.0-00010101000000-000000000000
	github.com/segmentio/kafka-go v0.4.9
)
