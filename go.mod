module github.com/ride4Low/notifier

go 1.25.3

require (
	github.com/bytedance/sonic v1.14.2
	github.com/gorilla/websocket v1.5.3
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/ride4Low/contracts v0.0.0-20251130095245-000000000000
	google.golang.org/grpc v1.77.0
)

require (
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/ride4Low/contracts => ../contracts
