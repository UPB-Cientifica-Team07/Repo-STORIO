# Requisitos de desarrollo - Monitoring Service

## 1. Requisitos

Para ejecutar y desarrollar el Monitoring Service se requiere instalar:

- Go 1.25 o superior.
- Protocol Buffers Compiler (`protoc`).
- Plugin `protoc-gen-go`.
- Plugin `protoc-gen-go-grpc`.
- Git.

---

## 2. Verificar Go

```bash
go version
```

El proyecto utiliza Go y las dependencias se gestionan mediante:

```text
go.mod
go.sum
```

Para descargar las dependencias:

```bash
go mod download
```

---

## 3. Instalar Protocol Buffers

En sistemas basados en Debian, Ubuntu o Kali Linux:

```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

Verificar:

```bash
protoc --version
```

---

## 4. Instalar los plugins de Go para Protocol Buffers

Instalar el generador de código Go:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Instalar el generador gRPC:

```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## 5. Configurar PATH

Los plugins instalados por Go se almacenan normalmente en:

```text
$(go env GOPATH)/bin
```

Agregar esta ruta al PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Para Zsh:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

Para Bash:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

---

## 6. Verificar herramientas

Ejecutar:

```bash
go version
protoc --version
which protoc-gen-go
which protoc-gen-go-grpc
```

Las cuatro herramientas deben estar disponibles antes de generar el código gRPC.

---

## 7. Clonar el proyecto

```bash
git clone <URL_DEL_REPOSITORIO>
```

Entrar al servicio:

```bash
cd Repo-STORIO/services/monitoring-service
```

Descargar las dependencias:

```bash
go mod download
```

---

## 8. Generar archivos gRPC

Los archivos generados se crean a partir del contrato:

```text
proto/monitoring.proto
```

Ejecutar:

```bash
protoc \
  --proto_path=proto \
  --go_out=generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=generated \
  --go-grpc_opt=paths=source_relative \
  proto/monitoring.proto
```

Esto genera:

```text
generated/
├── monitoring.pb.go
└── monitoring_grpc.pb.go
```

---

## 9. Ejecutar el servidor

Desde:

```text
services/monitoring-service/
```

Ejecutar:

```bash
go run ./cmd/server
```

El servicio inicia en:

```text
Puerto: 50051
Protocolo: gRPC
```

---

## 10. Ejecutar el cliente de prueba

En otra terminal:

```bash
go run ./cmd/client
```

El cliente envía una solicitud `ReportMetrics()` al Monitoring Service mediante gRPC.

---

## 11. Estructura principal

```text
monitoring-service/
├── cmd/
│   ├── client/
│   └── server/
│
├── generated/
│   ├── monitoring.pb.go
│   └── monitoring_grpc.pb.go
│
├── internal/
│   ├── grpc/
│   ├── model/
│   ├── repository/
│   └── service/
│
├── proto/
│   └── monitoring.proto
│
├── go.mod
├── go.sum
├── README.md
└── requirements.md
```

## 12. Dependencias del proyecto

Las dependencias específicas están definidas automáticamente en:

```text
go.mod
```

Los checksums de las dependencias se almacenan en:

```text
go.sum
```

Estos archivos deben mantenerse versionados en el repositorio.

No se deben subir directorios de caché de Go ni binarios compilados.