module github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service

go 1.25.0

replace github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service => ../monitoring-service

require (
	github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service v0.0.0
	google.golang.org/grpc v1.83.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
