package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"github.com/fsnotify/fsnotify"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	defaultServerAddress = "localhost:50055"
	defaultAuthURL       = "http://localhost:8081"

	defaultUsername = "tercero"
	defaultPassword = "123456"

	defaultDeviceID = "linux-client-001"

	defaultSyncDirectory = "./sync-home"

	stateFileName = ".sync-state.json"

	uploadChunkSize = 64 * 1024
)

type Config struct {
	ServerAddress string
	AuthURL       string

	Username string
	Password string

	DeviceID string

	SyncDirectory string
}

type StateEntry struct {
	FileID string `json:"file_id"`

	SHA256 string `json:"sha256,omitempty"`

	Size int64 `json:"size,omitempty"`

	ModifiedAt int64 `json:"modified_at,omitempty"`
}

// =====================================
// COMPATIBILIDAD STATE V1 -> V2
// =====================================
//
// Formato anterior:
//
// "ruta.txt": "uuid"
//
// Formato nuevo:
//
// "ruta.txt": {
//     "file_id": "...",
//     "sha256": "...",
//     "size": 123,
//     "modified_at": ...
// }
//
// =====================================

func (e *StateEntry) UnmarshalJSON(
	data []byte,
) error {

	if len(data) > 0 &&
		data[0] == '"' {

		var legacyFileID string

		if err := json.Unmarshal(
			data,
			&legacyFileID,
		); err != nil {

			return err
		}

		e.FileID =
			legacyFileID

		return nil
	}

	type stateEntryAlias StateEntry

	var decoded stateEntryAlias

	if err := json.Unmarshal(
		data,
		&decoded,
	); err != nil {

		return err
	}

	*e =
		StateEntry(
			decoded,
		)

	return nil
}

type StateStore struct {
	mu sync.RWMutex

	path string

	Files map[string]StateEntry `json:"files"`
}

type Client struct {
	config Config

	token  string
	userID string

	connection *grpc.ClientConn
	client     pb.SyncServiceClient

	state *StateStore

	suppressMu sync.Mutex

	suppressUntil map[string]time.Time

	debounceMu sync.Mutex

	debounce map[string]*time.Timer
}

// =====================================
// MAIN
// =====================================

func main() {

	config, err :=
		loadConfig()

	if err != nil {

		log.Fatalf(
			"Configuración inválida: %v",
			err,
		)
	}

	log.Println("===================================")
	log.Println(" FILE SYNC CLIENT - LINUX")
	log.Println("===================================")
	log.Println("Servidor:", config.ServerAddress)
	log.Println("Device:", config.DeviceID)
	log.Println("Directorio:", config.SyncDirectory)
	log.Println("===================================")

	err =
		os.MkdirAll(
			config.SyncDirectory,
			0750,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo crear directorio sincronizado: %v",
			err,
		)
	}

	statePath :=
		filepath.Join(
			config.SyncDirectory,
			stateFileName,
		)

	state, err :=
		loadState(
			statePath,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo cargar estado local: %v",
			err,
		)
	}

	authClient :=
		auth.NewClient(
			config.AuthURL,
		)

	loginResult, err :=
		authClient.Login(
			config.Username,
			config.Password,
		)

	if err != nil {

		log.Fatalf(
			"Login falló: %v",
			err,
		)
	}

	if !loginResult.Success {

		log.Fatalf(
			"Login rechazado: %s",
			loginResult.Message,
		)
	}

	if strings.TrimSpace(
		loginResult.Token,
	) == "" {

		log.Fatal(
			"Auth Service no devolvió token",
		)
	}

	log.Println(
		"Usuario autenticado:",
		loginResult.UserID,
	)

	connection, err :=
		grpc.NewClient(
			config.ServerAddress,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {

		log.Fatalf(
			"No se pudo conectar con Sync Service: %v",
			err,
		)
	}

	defer connection.Close()

	syncClient :=
		&Client{
			config: config,

			token: loginResult.Token,

			userID: loginResult.UserID,

			connection: connection,

			client: pb.NewSyncServiceClient(
				connection,
			),

			state: state,

			suppressUntil: make(
				map[string]time.Time,
			),

			debounce: make(
				map[string]*time.Timer,
			),
		}

	err =
		syncClient.authenticate()

	if err != nil {

		log.Fatalf(
			"Authenticate falló: %v",
			err,
		)
	}

	log.Println(
		"Identidad Sync validada",
	)

	// =====================================
	// RECONCILIACIÓN INICIAL
	// =====================================

	log.Println(
		"Ejecutando reconciliación inicial...",
	)

	err =
		syncClient.consumePendingChanges()

	if err != nil {

		log.Printf(
			"Advertencia consumiendo cambios pendientes: %v",
			err,
		)
	}

	err =
		syncClient.reconcileRemoteFiles()

	if err != nil {

		log.Fatalf(
			"Reconciliación remota falló: %v",
			err,
		)
	}

	err =
		syncClient.reconcileLocalFiles()

	if err != nil {

		log.Fatalf(
			"Reconciliación local falló: %v",
			err,
		)
	}

	log.Println(
		"Reconciliación inicial completada",
	)

	// =====================================
	// WATCHER FS
	// =====================================

	watcher, err :=
		fsnotify.NewWatcher()

	if err != nil {

		log.Fatalf(
			"No se pudo crear watcher: %v",
			err,
		)
	}

	defer watcher.Close()

	err =
		addWatchRecursive(
			watcher,
			config.SyncDirectory,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo vigilar árbol local: %v",
			err,
		)
	}

	// =====================================
	// WATCH REMOTO
	// =====================================

	go syncClient.watchRemoteChanges()

	// =====================================
	// WATCH LOCAL
	// =====================================

	go syncClient.watchLocalChanges(
		watcher,
	)

	log.Println("===================================")
	log.Println(" FILE SYNC ACTIVO")
	log.Println("===================================")
	log.Println("Cambios locales -> servidor")
	log.Println("Cambios servidor -> directorio local")
	log.Println("===================================")

	stop :=
		make(
			chan os.Signal,
			1,
		)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println(
		"Deteniendo File Sync Client...",
	)

	signal.Stop(
		stop,
	)

	log.Println(
		"File Sync Client detenido",
	)
}

// =====================================
// CONFIG
// =====================================

func loadConfig() (
	Config,
	error,
) {

	config :=
		Config{
			ServerAddress: getenv(
				"SYNC_SERVER",
				defaultServerAddress,
			),

			AuthURL: getenv(
				"SYNC_AUTH_URL",
				defaultAuthURL,
			),

			Username: getenv(
				"SYNC_USERNAME",
				defaultUsername,
			),

			Password: getenv(
				"SYNC_PASSWORD",
				defaultPassword,
			),

			DeviceID: getenv(
				"SYNC_DEVICE_ID",
				defaultDeviceID,
			),

			SyncDirectory: getenv(
				"SYNC_DIRECTORY",
				defaultSyncDirectory,
			),
		}

	absolutePath, err :=
		filepath.Abs(
			config.SyncDirectory,
		)

	if err != nil {

		return Config{},
			fmt.Errorf(
				"no se pudo resolver directorio Sync: %w",
				err,
			)
	}

	config.SyncDirectory =
		filepath.Clean(
			absolutePath,
		)

	if strings.TrimSpace(
		config.DeviceID,
	) == "" {

		return Config{},
			fmt.Errorf(
				"SYNC_DEVICE_ID es obligatorio",
			)
	}

	return config, nil
}

func getenv(
	key string,
	fallback string,
) string {

	value :=
		strings.TrimSpace(
			os.Getenv(
				key,
			),
		)

	if value == "" {
		return fallback
	}

	return value
}

// =====================================
// AUTH CONTEXT
// =====================================

func (c *Client) authenticatedContext(
	timeout time.Duration,
) (
	context.Context,
	context.CancelFunc,
) {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			timeout,
		)

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+c.token,
			),
		)

	return ctx, cancel
}

func (c *Client) authenticate() error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.Authenticate(
			ctx,
			&pb.AuthenticateRequest{
				Token: c.token,

				DeviceId: c.config.DeviceID,
			},
		)

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	if response.UserId !=
		c.userID {

		return fmt.Errorf(
			"identidad inconsistente: login=%s sync=%s",
			c.userID,
			response.UserId,
		)
	}

	return nil
}

// =====================================
// STATE
// =====================================

func loadState(
	path string,
) (
	*StateStore,
	error,
) {

	state :=
		&StateStore{
			path: path,

			Files: make(
				map[string]StateEntry,
			),
		}

	content, err :=
		os.ReadFile(
			path,
		)

	if os.IsNotExist(
		err,
	) {

		return state, nil
	}

	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return state, nil
	}

	err =
		json.Unmarshal(
			content,
			state,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"estado local inválido: %w",
				err,
			)
	}

	if state.Files == nil {

		state.Files =
			make(
				map[string]StateEntry,
			)
	}

	state.path =
		path

	return state, nil
}

// =====================================
// FINGERPRINT SHA-256
// =====================================

func fingerprintFile(
	fullPath string,
) (
	string,
	int64,
	int64,
	error,
) {

	file, err :=
		os.Open(
			fullPath,
		)

	if err != nil {

		return "",
			0,
			0,
			err
	}

	defer file.Close()

	hasher :=
		sha256.New()

	_, err =
		io.Copy(
			hasher,
			file,
		)

	if err != nil {

		return "",
			0,
			0,
			err
	}

	info, err :=
		file.Stat()

	if err != nil {

		return "",
			0,
			0,
			err
	}

	hash :=
		fmt.Sprintf(
			"%x",
			hasher.Sum(nil),
		)

	return hash,
		info.Size(),
		info.ModTime().UnixNano(),
		nil
}

func (s *StateStore) save() error {

	s.mu.RLock()

	copyFiles :=
		make(
			map[string]StateEntry,
			len(s.Files),
		)

	for relativePath, entry := range s.Files {

		copyFiles[relativePath] =
			entry
	}

	s.mu.RUnlock()

	payload, err :=
		json.MarshalIndent(
			struct {
				Files map[string]StateEntry `json:"files"`
			}{
				Files: copyFiles,
			},
			"",
			"  ",
		)

	if err != nil {
		return err
	}

	temporary :=
		s.path +
			".tmp"

	err =
		os.WriteFile(
			temporary,
			payload,
			0600,
		)

	if err != nil {
		return err
	}

	return os.Rename(
		temporary,
		s.path,
	)
}

func (s *StateStore) get(
	relativePath string,
) (
	string,
	bool,
) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok :=
		s.Files[relativePath]

	if !ok {
		return "", false
	}

	return entry.FileID, true
}

func (s *StateStore) getEntry(
	relativePath string,
) (
	StateEntry,
	bool,
) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok :=
		s.Files[relativePath]

	return entry, ok
}

func (s *StateStore) snapshot() map[string]StateEntry {

	s.mu.RLock()
	defer s.mu.RUnlock()

	result :=
		make(
			map[string]StateEntry,
			len(s.Files),
		)

	for relativePath, entry := range s.Files {

		result[relativePath] =
			entry
	}

	return result
}

func (s *StateStore) findPathByFileID(
	fileID string,
) (
	string,
	bool,
) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	for relativePath, entry := range s.Files {

		if entry.FileID ==
			fileID {

			return relativePath, true
		}
	}

	return "", false
}

// =====================================
// SET SIMPLE
// =====================================
//
// Actualiza solamente file_id y conserva
// SHA/size/mtime existentes.
//
// =====================================

func (s *StateStore) set(
	relativePath string,
	fileID string,
) error {

	s.mu.Lock()

	entry :=
		s.Files[relativePath]

	entry.FileID =
		fileID

	s.Files[relativePath] =
		entry

	s.mu.Unlock()

	return s.save()
}

// =====================================
// SET DESDE ARCHIVO LOCAL
// =====================================

func (s *StateStore) setLocal(
	relativePath string,
	fileID string,
	fullPath string,
) error {

	hash,
		size,
		modifiedAt,
		err :=
		fingerprintFile(
			fullPath,
		)

	if err != nil {
		return err
	}

	s.mu.Lock()

	s.Files[relativePath] =
		StateEntry{
			FileID: fileID,

			SHA256: hash,

			Size: size,

			ModifiedAt: modifiedAt,
		}

	s.mu.Unlock()

	return s.save()
}

func (s *StateStore) delete(
	relativePath string,
) error {

	s.mu.Lock()

	delete(
		s.Files,
		relativePath,
	)

	s.mu.Unlock()

	return s.save()
}

// =====================================
// PATH SEGURA
// =====================================

func (c *Client) relativePath(
	fullPath string,
) (
	string,
	error,
) {

	relativePath, err :=
		filepath.Rel(
			c.config.SyncDirectory,
			fullPath,
		)

	if err != nil {
		return "", err
	}

	if relativePath == "." {

		return "", nil
	}

	relativePath =
		filepath.ToSlash(
			relativePath,
		)

	if relativePath == stateFileName ||
		relativePath == stateFileName+".tmp" {

		return "", nil
	}

	if relativePath == ".." ||
		strings.HasPrefix(
			relativePath,
			"../",
		) {

		return "",
			fmt.Errorf(
				"ruta fuera del directorio sincronizado",
			)
	}

	return relativePath, nil
}

func safeLocalPath(
	root string,
	relativePath string,
) (
	string,
	error,
) {

	relativePath =
		strings.TrimSpace(
			relativePath,
		)

	relativePath =
		strings.ReplaceAll(
			relativePath,
			"\\",
			"/",
		)

	if relativePath == "" {

		return "",
			fmt.Errorf(
				"relative_path vacío",
			)
	}

	if strings.HasPrefix(
		relativePath,
		"/",
	) {

		return "",
			fmt.Errorf(
				"relative_path absoluto no permitido",
			)
	}

	if len(relativePath) >= 3 &&
		relativePath[1] == ':' &&
		relativePath[2] == '/' {

		return "",
			fmt.Errorf(
				"ruta Windows absoluta no permitida",
			)
	}

	clean :=
		filepath.Clean(
			filepath.FromSlash(
				relativePath,
			),
		)

	if clean == ".." ||
		strings.HasPrefix(
			clean,
			".."+string(
				os.PathSeparator,
			),
		) {

		return "",
			fmt.Errorf(
				"path traversal no permitido",
			)
	}

	full :=
		filepath.Join(
			root,
			clean,
		)

	relativeCheck, err :=
		filepath.Rel(
			root,
			full,
		)

	if err != nil {
		return "", err
	}

	if relativeCheck == ".." ||
		strings.HasPrefix(
			relativeCheck,
			".."+string(
				os.PathSeparator,
			),
		) {

		return "",
			fmt.Errorf(
				"ruta final fuera del Home local",
			)
	}

	return full, nil
}

// =====================================
// SUPRESIÓN DE EVENTOS LOCALES
// =====================================

func (c *Client) suppress(
	relativePath string,
	duration time.Duration,
) {

	c.suppressMu.Lock()

	c.suppressUntil[relativePath] =
		time.Now().
			Add(
				duration,
			)

	c.suppressMu.Unlock()
}

func (c *Client) isSuppressed(
	relativePath string,
) bool {

	c.suppressMu.Lock()
	defer c.suppressMu.Unlock()

	until, ok :=
		c.suppressUntil[relativePath]

	if !ok {
		return false
	}

	if time.Now().
		After(
			until,
		) {

		delete(
			c.suppressUntil,
			relativePath,
		)

		return false
	}

	return true
}

// =====================================
// UPLOAD
// =====================================

func (c *Client) uploadLocalFile(
	fullPath string,
	relativePath string,
) error {

	content, err :=
		os.ReadFile(
			fullPath,
		)

	if err != nil {
		return err
	}

	fileType :=
		mime.TypeByExtension(
			filepath.Ext(
				fullPath,
			),
		)

	if fileType == "" {

		fileType =
			"application/octet-stream"
	}

	ctx, cancel :=
		c.authenticatedContext(
			60 * time.Second,
		)

	defer cancel()

	stream, err :=
		c.client.Upload(
			ctx,
		)

	if err != nil {
		return err
	}

	fileName :=
		filepath.Base(
			fullPath,
		)

	err =
		stream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						UserId: c.userID,

						DeviceId: c.config.DeviceID,

						FileName: fileName,

						FileType: fileType,

						Size: int64(
							len(content),
						),

						RelativePath: relativePath,
					},
				},
			},
		)

	if err != nil {
		return err
	}

	for start := 0; start < len(content); start += uploadChunkSize {

		end :=
			start +
				uploadChunkSize

		if end >
			len(content) {

			end =
				len(content)
		}

		err =
			stream.Send(
				&pb.UploadRequest{
					Data: &pb.UploadRequest_Chunk{
						Chunk: content[start:end],
					},
				},
			)

		if err != nil {
			return err
		}
	}

	response, err :=
		stream.CloseAndRecv()

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	err =
		c.state.setLocal(
			relativePath,
			response.FileId,
			fullPath,
		)

	if err != nil {
		return err
	}

	log.Printf(
		"UPLOAD | %s | fileID=%s",
		relativePath,
		response.FileId,
	)

	return nil
}

// =====================================
// UPDATE
// =====================================

func (c *Client) updateLocalFile(
	fullPath string,
	relativePath string,
	fileID string,
) error {

	content, err :=
		os.ReadFile(
			fullPath,
		)

	if err != nil {
		return err
	}

	ctx, cancel :=
		c.authenticatedContext(
			60 * time.Second,
		)

	defer cancel()

	response, err :=
		c.client.UpdateFile(
			ctx,
			&pb.UpdateFileRequest{
				UserId: c.userID,

				DeviceId: c.config.DeviceID,

				FileId: fileID,

				Content: content,
			},
		)

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	err =
		c.state.setLocal(
			relativePath,
			fileID,
			fullPath,
		)

	if err != nil {
		return err
	}

	log.Printf(
		"UPDATE | %s | fileID=%s | version=%d",
		relativePath,
		fileID,
		response.Version,
	)

	return nil
}

// =====================================
// DELETE
// =====================================

func (c *Client) deleteLocalFile(
	relativePath string,
	fileID string,
) error {

	ctx, cancel :=
		c.authenticatedContext(
			30 * time.Second,
		)

	defer cancel()

	response, err :=
		c.client.DeleteFile(
			ctx,
			&pb.DeleteFileRequest{
				UserId: c.userID,

				DeviceId: c.config.DeviceID,

				FileId: fileID,
			},
		)

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	err =
		c.state.delete(
			relativePath,
		)

	if err != nil {
		return err
	}

	log.Printf(
		"DELETE | %s | fileID=%s",
		relativePath,
		fileID,
	)

	return nil
}

// =====================================
// DOWNLOAD
// =====================================

func (c *Client) downloadFile(
	fileID string,
) (
	*pb.FileMetadata,
	[]byte,
	error,
) {

	ctx, cancel :=
		c.authenticatedContext(
			60 * time.Second,
		)

	defer cancel()

	stream, err :=
		c.client.Download(
			ctx,
			&pb.DownloadRequest{
				UserId: c.userID,

				FileId: fileID,
			},
		)

	if err != nil {

		return nil,
			nil,
			err
	}

	var (
		fileMetadata *pb.FileMetadata
		buffer       bytes.Buffer
	)

	for {

		response, recvErr :=
			stream.Recv()

		if recvErr == io.EOF {
			break
		}

		if recvErr != nil {

			return nil,
				nil,
				recvErr
		}

		if response.GetMetadata() != nil {

			fileMetadata =
				response.GetMetadata()

			continue
		}

		if response.GetChunk() != nil {

			_, _ =
				buffer.Write(
					response.GetChunk(),
				)
		}
	}

	if fileMetadata == nil {

		return nil,
			nil,
			fmt.Errorf(
				"Download no devolvió metadata",
			)
	}

	return fileMetadata,
		buffer.Bytes(),
		nil
}

// =====================================
// APLICAR CAMBIO REMOTO
// =====================================

func (c *Client) applyRemoteChange(
	change *pb.FileChange,
) error {

	if change == nil {
		return nil
	}

	switch change.Type {

	case "FILE_CREATED",
		"FILE_CHANGED":

		metadataResponse,
			content,
			err :=
			c.downloadFile(
				change.FileId,
			)

		if err != nil {
			return err
		}

		relativePath :=
			metadataResponse.RelativePath

		if relativePath == "" {

			relativePath =
				change.RelativePath
		}

		fullPath, err :=
			safeLocalPath(
				c.config.SyncDirectory,
				relativePath,
			)

		if err != nil {
			return err
		}

		err =
			os.MkdirAll(
				filepath.Dir(
					fullPath,
				),
				0750,
			)

		if err != nil {
			return err
		}

		c.suppress(
			relativePath,
			2*time.Second,
		)

		temporary :=
			fullPath +
				".sync-tmp"

		err =
			os.WriteFile(
				temporary,
				content,
				0640,
			)

		if err != nil {
			return err
		}

		err =
			os.Rename(
				temporary,
				fullPath,
			)

		if err != nil {
			return err
		}

		err =
			c.state.setLocal(
				relativePath,
				change.FileId,
				fullPath,
			)

		if err != nil {
			return err
		}

		log.Printf(
			"REMOTE %s | %s | version=%d",
			change.Type,
			relativePath,
			change.Version,
		)

	case "FILE_DELETED":

		relativePath :=
			change.RelativePath

		if relativePath == "" {

			pathFromState, ok :=
				c.state.findPathByFileID(
					change.FileId,
				)

			if ok {

				relativePath =
					pathFromState
			}
		}

		if relativePath == "" {

			return nil
		}

		fullPath, err :=
			safeLocalPath(
				c.config.SyncDirectory,
				relativePath,
			)

		if err != nil {
			return err
		}

		c.suppress(
			relativePath,
			2*time.Second,
		)

		err =
			os.Remove(
				fullPath,
			)

		if err != nil &&
			!os.IsNotExist(
				err,
			) {

			return err
		}

		err =
			c.state.delete(
				relativePath,
			)

		if err != nil {
			return err
		}

		log.Printf(
			"REMOTE DELETE | %s | version=%d",
			relativePath,
			change.Version,
		)
	}

	return nil
}

// =====================================
// CONSUMIR CAMBIOS PENDIENTES
// =====================================

func (c *Client) consumePendingChanges() error {

	ctx, cancel :=
		c.authenticatedContext(
			30 * time.Second,
		)

	defer cancel()

	response, err :=
		c.client.Sync(
			ctx,
			&pb.SyncRequest{
				UserId: c.userID,

				DeviceId: c.config.DeviceID,
			},
		)

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	for _, change := range response.Changes {

		err =
			c.applyRemoteChange(
				change,
			)

		if err != nil {

			log.Printf(
				"Advertencia aplicando cambio %s: %v",
				change.FileId,
				err,
			)
		}
	}

	return nil
}

// =====================================
// RECONCILIAR REMOTO
// =====================================
//
// ListFiles representa el estado actual.
//
// Esto permite reconstruir .sync-state.json
// aunque el archivo local de estado se pierda.
//
// =====================================

func (c *Client) reconcileRemoteFiles() error {

	ctx, cancel :=
		c.authenticatedContext(
			30 * time.Second,
		)

	defer cancel()

	response, err :=
		c.client.ListFiles(
			ctx,
			&pb.ListFilesRequest{
				UserId: c.userID,
			},
		)

	if err != nil {
		return err
	}

	if !response.Success {

		return fmt.Errorf(
			"%s",
			response.Message,
		)
	}

	for _, file := range response.Files {

		if file == nil ||
			file.RelativePath == "" {

			continue
		}

		fullPath, err :=
			safeLocalPath(
				c.config.SyncDirectory,
				file.RelativePath,
			)

		if err != nil {

			log.Printf(
				"Ruta remota ignorada: %v",
				err,
			)

			continue
		}

		stateEntry, tracked :=
			c.state.getEntry(
				file.RelativePath,
			)

		_, statErr :=
			os.Stat(
				fullPath,
			)

		// =====================================
		// ARCHIVO YA CONOCIDO POR ESTE CLIENTE
		// =====================================
		//
		// Si falta localmente NO lo restauramos aquí.
		//
		// reconcileLocalFiles interpretará esto
		// como un DELETE realizado mientras el
		// cliente estaba apagado.
		//
		// =====================================

		if tracked {

			if statErr == nil {

				if stateEntry.FileID !=
					file.FileId {

					err =
						c.state.set(
							file.RelativePath,
							file.FileId,
						)

					if err != nil {
						return err
					}
				}

				continue
			}

			if os.IsNotExist(
				statErr,
			) {

				continue
			}

			return statErr
		}

		// =====================================
		// ARCHIVO REMOTO NUEVO PARA EL CLIENTE
		// =====================================

		meta,
			content,
			downloadErr :=
			c.downloadFile(
				file.FileId,
			)

		if downloadErr != nil {
			return downloadErr
		}

		err =
			os.MkdirAll(
				filepath.Dir(
					fullPath,
				),
				0750,
			)

		if err != nil {
			return err
		}

		err =
			os.WriteFile(
				fullPath,
				content,
				0640,
			)

		if err != nil {
			return err
		}

		err =
			c.state.setLocal(
				file.RelativePath,
				file.FileId,
				fullPath,
			)

		if err != nil {
			return err
		}

		log.Printf(
			"RECONCILE REMOTE | %s | version=%d",
			meta.RelativePath,
			meta.Version,
		)
	}

	return nil
}

// =====================================
// RECONCILIAR LOCAL
// =====================================
//
// Detecta también cambios ocurridos mientras
// el watcher estuvo apagado:
//
// NEW      -> Upload
// MODIFIED -> UpdateFile
// DELETED  -> DeleteFile
//
// =====================================

func (c *Client) reconcileLocalFiles() error {

	seen :=
		make(
			map[string]bool,
		)

	err :=
		filepath.WalkDir(
			c.config.SyncDirectory,
			func(
				currentPath string,
				entry os.DirEntry,
				walkErr error,
			) error {

				if walkErr != nil {
					return walkErr
				}

				if entry.IsDir() {
					return nil
				}

				relativePath, err :=
					c.relativePath(
						currentPath,
					)

				if err != nil {
					return err
				}

				if relativePath == "" {
					return nil
				}

				seen[relativePath] =
					true

				stateEntry, exists :=
					c.state.getEntry(
						relativePath,
					)

				// =================================
				// ARCHIVO NUEVO
				// =================================

				if !exists {

					log.Printf(
						"OFFLINE CREATE detectado | %s",
						relativePath,
					)

					return c.uploadLocalFile(
						currentPath,
						relativePath,
					)
				}

				// =================================
				// MIGRACIÓN DEL STATE V1
				// =================================

				if stateEntry.SHA256 == "" {

					err =
						c.state.setLocal(
							relativePath,
							stateEntry.FileID,
							currentPath,
						)

					if err != nil {
						return err
					}

					log.Printf(
						"STATE SHA-256 inicializado | %s",
						relativePath,
					)

					return nil
				}

				hash,
					size,
					modifiedAt,
					err :=
					fingerprintFile(
						currentPath,
					)

				if err != nil {
					return err
				}

				// =================================
				// ARCHIVO MODIFICADO
				// =================================

				if hash !=
					stateEntry.SHA256 ||
					size !=
						stateEntry.Size {

					log.Printf(
						"OFFLINE MODIFY detectado | %s",
						relativePath,
					)

					return c.updateLocalFile(
						currentPath,
						relativePath,
						stateEntry.FileID,
					)
				}

				// Contenido igual, pero actualizamos
				// mtime si cambió.

				if modifiedAt !=
					stateEntry.ModifiedAt {

					return c.state.setLocal(
						relativePath,
						stateEntry.FileID,
						currentPath,
					)
				}

				return nil
			},
		)

	if err != nil {
		return err
	}

	// =====================================
	// ARCHIVOS ELIMINADOS DURANTE OFFLINE
	// =====================================

	snapshot :=
		c.state.snapshot()

	for relativePath, stateEntry := range snapshot {

		if seen[relativePath] {
			continue
		}

		if strings.TrimSpace(
			stateEntry.FileID,
		) == "" {

			continue
		}

		fullPath, err :=
			safeLocalPath(
				c.config.SyncDirectory,
				relativePath,
			)

		if err != nil {
			return err
		}

		_, statErr :=
			os.Stat(
				fullPath,
			)

		if statErr == nil {
			continue
		}

		if !os.IsNotExist(
			statErr,
		) {

			return statErr
		}

		log.Printf(
			"OFFLINE DELETE detectado | %s",
			relativePath,
		)

		err =
			c.deleteLocalFile(
				relativePath,
				stateEntry.FileID,
			)

		if err != nil {
			return err
		}
	}

	return nil
}

// =====================================
// WATCH LOCAL
// =====================================

func (c *Client) watchLocalChanges(
	watcher *fsnotify.Watcher,
) {

	for {

		select {

		case event, ok :=
			<-watcher.Events:

			if !ok {
				return
			}

			c.handleLocalEvent(
				watcher,
				event,
			)

		case watcherErr, ok :=
			<-watcher.Errors:

			if !ok {
				return
			}

			log.Printf(
				"Watcher error: %v",
				watcherErr,
			)
		}
	}
}

func (c *Client) handleLocalEvent(
	watcher *fsnotify.Watcher,
	event fsnotify.Event,
) {

	relativePath, err :=
		c.relativePath(
			event.Name,
		)

	if err != nil {

		log.Printf(
			"Evento inválido: %v",
			err,
		)

		return
	}

	if relativePath == "" {
		return
	}

	if c.isSuppressed(
		relativePath,
	) {

		return
	}

	if event.Op&
		fsnotify.Create != 0 {

		info, statErr :=
			os.Stat(
				event.Name,
			)

		if statErr == nil &&
			info.IsDir() {

			err =
				addWatchRecursive(
					watcher,
					event.Name,
				)

			if err != nil {

				log.Printf(
					"No se pudo vigilar nuevo directorio: %v",
					err,
				)
			}

			return
		}

		c.scheduleLocalUpsert(
			event.Name,
			relativePath,
		)
	}

	if event.Op&
		fsnotify.Write != 0 {

		c.scheduleLocalUpsert(
			event.Name,
			relativePath,
		)
	}

	if event.Op&fsnotify.Remove != 0 ||
		event.Op&fsnotify.Rename != 0 {

		fileID, exists :=
			c.state.get(
				relativePath,
			)

		if !exists {
			return
		}

		go func() {

			err :=
				c.deleteLocalFile(
					relativePath,
					fileID,
				)

			if err != nil {

				log.Printf(
					"DELETE local falló %s: %v",
					relativePath,
					err,
				)
			}
		}()
	}
}

// =====================================
// DEBOUNCE
// =====================================

func (c *Client) scheduleLocalUpsert(
	fullPath string,
	relativePath string,
) {

	c.debounceMu.Lock()

	if currentTimer, ok :=
		c.debounce[relativePath]; ok {

		currentTimer.Stop()
	}

	c.debounce[relativePath] =
		time.AfterFunc(
			500*time.Millisecond,
			func() {

				c.debounceMu.Lock()

				delete(
					c.debounce,
					relativePath,
				)

				c.debounceMu.Unlock()

				c.syncLocalUpsert(
					fullPath,
					relativePath,
				)
			},
		)

	c.debounceMu.Unlock()
}

func (c *Client) syncLocalUpsert(
	fullPath string,
	relativePath string,
) {

	info, err :=
		os.Stat(
			fullPath,
		)

	if err != nil {
		return
	}

	if info.IsDir() {
		return
	}

	if c.isSuppressed(
		relativePath,
	) {

		return
	}

	stateEntry, exists :=
		c.state.getEntry(
			relativePath,
		)

	// =====================================
	// ARCHIVO YA REGISTRADO
	// =====================================

	if exists {

		hash,
			size,
			modifiedAt,
			err :=
			fingerprintFile(
				fullPath,
			)

		if err != nil {

			log.Printf(
				"Fingerprint local falló %s: %v",
				relativePath,
				err,
			)

			return
		}

		// =================================
		// CONTENIDO REALMENTE IGUAL
		// =================================
		//
		// fsnotify puede emitir múltiples
		// WRITE para una sola operación.
		//
		// No generamos una nueva versión
		// si SHA-256 y tamaño son iguales.
		//
		// =================================

		if stateEntry.SHA256 != "" &&
			hash ==
				stateEntry.SHA256 &&
			size ==
				stateEntry.Size {

			if modifiedAt !=
				stateEntry.ModifiedAt {

				err =
					c.state.setLocal(
						relativePath,
						stateEntry.FileID,
						fullPath,
					)

				if err != nil {

					log.Printf(
						"Actualización de estado local falló %s: %v",
						relativePath,
						err,
					)
				}
			}

			return
		}

		err =
			c.updateLocalFile(
				fullPath,
				relativePath,
				stateEntry.FileID,
			)

		if err != nil {

			log.Printf(
				"UPDATE local falló %s: %v",
				relativePath,
				err,
			)
		}

		return
	}

	// =====================================
	// ARCHIVO NUEVO
	// =====================================

	err =
		c.uploadLocalFile(
			fullPath,
			relativePath,
		)

	if err != nil {

		log.Printf(
			"UPLOAD local falló %s: %v",
			relativePath,
			err,
		)
	}
}

// =====================================
// WATCH REMOTO
// =====================================

func (c *Client) watchRemoteChanges() {

	for {

		err :=
			c.runRemoteWatch()

		if err != nil {

			log.Printf(
				"WatchChanges desconectado: %v",
				err,
			)
		}

		time.Sleep(
			3 * time.Second,
		)

		log.Println(
			"Reconectando WatchChanges...",
		)
	}
}

func (c *Client) runRemoteWatch() error {

	ctx :=
		metadata.NewOutgoingContext(
			context.Background(),
			metadata.Pairs(
				"authorization",
				"Bearer "+c.token,
			),
		)

	stream, err :=
		c.client.WatchChanges(
			ctx,
			&pb.WatchChangesRequest{
				UserId: c.userID,

				DeviceId: c.config.DeviceID,
			},
		)

	if err != nil {
		return err
	}

	for {

		change, err :=
			stream.Recv()

		if err != nil {
			return err
		}

		if change == nil {
			continue
		}

		if change.OriginDeviceId ==
			c.config.DeviceID {

			continue
		}

		err =
			c.applyRemoteChange(
				change,
			)

		if err != nil {

			log.Printf(
				"No se pudo aplicar cambio remoto %s/%s: %v",
				change.Type,
				change.FileId,
				err,
			)
		}
	}
}

// =====================================
// WATCH DIRECTORIOS
// =====================================

func addWatchRecursive(
	watcher *fsnotify.Watcher,
	root string,
) error {

	return filepath.WalkDir(
		root,
		func(
			currentPath string,
			entry os.DirEntry,
			walkErr error,
		) error {

			if walkErr != nil {
				return walkErr
			}

			if !entry.IsDir() {
				return nil
			}

			return watcher.Add(
				currentPath,
			)
		},
	)
}
