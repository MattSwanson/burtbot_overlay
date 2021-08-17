module github.com/MattSwanson/burtbot_overlay

go 1.16

require (
	cloud.google.com/go v0.86.0
	github.com/MattSwanson/ant-go v0.0.0
	github.com/MattSwanson/raylib-go v0.0.8
	github.com/google/gousb v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/ojrac/opensimplex-go v1.0.2
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	google.golang.org/genproto v0.0.0-20210701133433-6b8dcf568a95
)

replace github.com/MattSwanson/ant-go => ../ant-go
