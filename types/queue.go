package types

import "arkis_test/processor"

type QueuePair struct {
	Name   string
	Input  processor.Queue
	Output processor.Queue
}
