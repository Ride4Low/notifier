module github.com/sithu-go/ride-share/notifier

go 1.25.3

require (
	github.com/gorilla/websocket v1.5.3
	github.com/sithu-go/ride-share/contracts v0.0.0-00010101000000-000000000000
)

replace github.com/sithu-go/ride-share/contracts => ../contracts
