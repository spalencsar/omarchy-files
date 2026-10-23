package ipc

import dankipc "github.com/duxlabs-dev/dux-go/ipc"

const APIVersion = 1

type (
	Request    = dankipc.Request
	ConnWriter = dankipc.ConnWriter
	Subscriber = dankipc.Subscriber
	EventBus   = dankipc.EventBus
)

var (
	NewEventBus   = dankipc.NewEventBus
	NewConnWriter = dankipc.NewConnWriter
	Dial          = dankipc.Dial
)

func FindRunningSocket() (string, error) { return dankipc.FindRunningSocket("omarchy-files") }
