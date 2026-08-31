module github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service

go 1.25.13

require (
	github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service v0.0.0-20260823061311-9857a2c5b524
	github.com/fsnotify/fsnotify v1.10.1
	github.com/jackc/pgx/v5 v5.10.0
	google.golang.org/grpc v1.83.1
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

require (
	github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service v0.0.0
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
)

replace github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service => ../monitoring-service
