package scientific

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MaxSourceBytes  = 512 * 1024
	MaxDatasetBytes = 2 * 1024 * 1024
	MaxProcesses    = 4
)

var (
	safeWorkerPattern = regexp.MustCompile(
		`^[A-Za-z0-9._-]+@[A-Za-z0-9.:-]+$`,
	)

	safeRemotePathPattern = regexp.MustCompile(
		`^/[A-Za-z0-9._/-]+$`,
	)
)

type Result struct {
	JobID         string
	Success       bool
	ExitCode      int32
	LauncherNode  string
	DurationMs    int64
	Output        string
	SourceSHA256  string
	DatasetSHA256 string
}

type Executor struct {
	workers         []string
	remoteJobRoot   string
	coordinatorHost string
	coordinatorPort string
	javaClasspath   string
}

func NewExecutorFromEnv() (
	*Executor,
	error,
) {

	workersRaw :=
		getEnv(
			"GRID_HPC_WORKERS",
			"info@10.152.172.142,info@10.152.172.143",
		)

	parts :=
		strings.Split(
			workersRaw,
			",",
		)

	workers :=
		make(
			[]string,
			0,
			len(parts),
		)

	for _, value := range parts {

		worker :=
			strings.TrimSpace(
				value,
			)

		if worker == "" {
			continue
		}

		if !safeWorkerPattern.MatchString(
			worker,
		) {

			return nil,
				fmt.Errorf(
					"GRID_HPC_WORKERS contiene un destino inválido",
				)
		}

		workers =
			append(
				workers,
				worker,
			)
	}

	if len(workers) == 0 {

		return nil,
			fmt.Errorf(
				"GRID_HPC_WORKERS no contiene workers",
			)
	}

	remoteJobRoot :=
		getEnv(
			"GRID_HPC_JOB_ROOT",
			"/home/info/upb-hpc/jobs",
		)

	if !safeRemotePathPattern.MatchString(
		remoteJobRoot,
	) {

		return nil,
			fmt.Errorf(
				"GRID_HPC_JOB_ROOT inválido",
			)
	}

	coordinatorHost :=
		strings.TrimSpace(
			os.Getenv(
				"GRID_HPC_COORDINATOR_HOST",
			),
		)

	if coordinatorHost == "" {

		coordinatorHost =
			strings.TrimSpace(
				os.Getenv(
					"COORDINATOR_IP",
				),
			)
	}

	if coordinatorHost == "" {
		coordinatorHost = "127.0.0.1"
	}

	coordinatorPort :=
		strings.TrimSpace(
			os.Getenv(
				"GRID_HPC_COORDINATOR_PORT",
			),
		)

	if coordinatorPort == "" {

		coordinatorPort =
			strings.TrimSpace(
				os.Getenv(
					"HPC_RMI_PORT",
				),
			)
	}

	if coordinatorPort == "" {
		coordinatorPort = "1100"
	}

	port, err :=
		strconv.Atoi(
			coordinatorPort,
		)

	if err != nil ||
		port < 1 ||
		port > 65535 {

		return nil,
			fmt.Errorf(
				"puerto HPC inválido",
			)
	}

	javaClasspath :=
		getEnv(
			"GRID_HPC_JAVA_CLASSPATH",
			"services/hpc-service/bin",
		)

	return &Executor{
		workers:         workers,
		remoteJobRoot:   remoteJobRoot,
		coordinatorHost: coordinatorHost,
		coordinatorPort: coordinatorPort,
		javaClasspath:   javaClasspath,
	}, nil
}

func ValidateRequest(
	sourceCode []byte,
	dataset []byte,
	processes int32,
) error {

	if len(sourceCode) == 0 {

		return fmt.Errorf(
			"source_code es obligatorio",
		)
	}

	if len(sourceCode) >
		MaxSourceBytes {

		return fmt.Errorf(
			"source_code excede %d bytes",
			MaxSourceBytes,
		)
	}

	if len(dataset) == 0 {

		return fmt.Errorf(
			"dataset es obligatorio",
		)
	}

	if len(dataset) >
		MaxDatasetBytes {

		return fmt.Errorf(
			"dataset excede %d bytes",
			MaxDatasetBytes,
		)
	}

	if processes < 1 ||
		processes > MaxProcesses {

		return fmt.Errorf(
			"processes debe estar entre 1 y %d",
			MaxProcesses,
		)
	}

	return nil
}

func (e *Executor) Submit(
	ctx context.Context,
	token string,
	sourceCode []byte,
	dataset []byte,
	processes int32,
) (*Result, error) {

	if err :=
		ValidateRequest(
			sourceCode,
			dataset,
			processes,
		); err != nil {

		return nil, err
	}

	token =
		strings.TrimSpace(
			token,
		)

	if token == "" {

		return nil,
			fmt.Errorf(
				"token HPC vacío",
			)
	}

	jobID, err :=
		newUUID()

	if err != nil {
		return nil, err
	}

	sourceHash :=
		sha256Hex(
			sourceCode,
		)

	datasetHash :=
		sha256Hex(
			dataset,
		)

	localDirectory, err :=
		os.MkdirTemp(
			"",
			"upb-grid-scientific-"+jobID+"-",
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo crear workspace local: %w",
				err,
			)
	}

	defer os.RemoveAll(
		localDirectory,
	)

	sourcePath :=
		filepath.Join(
			localDirectory,
			"source.c",
		)

	datasetPath :=
		filepath.Join(
			localDirectory,
			"dataset",
		)

	if err :=
		os.WriteFile(
			sourcePath,
			sourceCode,
			0600,
		); err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo guardar source.c: %w",
				err,
			)
	}

	if err :=
		os.WriteFile(
			datasetPath,
			dataset,
			0600,
		); err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo guardar dataset: %w",
				err,
			)
	}

	for _, worker := range e.workers {

		if err :=
			e.stageWorker(
				ctx,
				worker,
				jobID,
				sourcePath,
				datasetPath,
				sourceHash,
				datasetHash,
			); err != nil {

			e.cleanupRemote(
				jobID,
			)

			return nil,
				fmt.Errorf(
					"staging en %s: %w",
					worker,
					err,
				)
		}
	}

	result, err :=
		e.executeHPC(
			ctx,
			token,
			jobID,
			processes,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"ejecución HPC: %w",
				err,
			)
	}

	result.SourceSHA256 =
		sourceHash

	result.DatasetSHA256 =
		datasetHash

	return result, nil
}

func (e *Executor) stageWorker(
	ctx context.Context,
	worker string,
	jobID string,
	sourcePath string,
	datasetPath string,
	sourceHash string,
	datasetHash string,
) error {

	remoteDirectory :=
		e.remoteJobRoot +
			"/" +
			jobID

	mkdirScript := `set -eu
JOB_DIR="$1"
mkdir -p "$JOB_DIR"
chmod 700 "$JOB_DIR"
`

	if _, err :=
		e.runSSH(
			ctx,
			10*time.Second,
			worker,
			mkdirScript,
			remoteDirectory,
		); err != nil {

		return err
	}

	destination :=
		worker +
			":" +
			remoteDirectory +
			"/"

	if _, err :=
		runCommand(
			ctx,
			20*time.Second,
			"",
			"scp",
			"-q",
			"-o",
			"BatchMode=yes",
			"-o",
			"ConnectTimeout=5",
			sourcePath,
			datasetPath,
			destination,
		); err != nil {

		return err
	}

	compileScript := `set -eu

JOB_DIR="$1"
EXPECTED_SOURCE="$2"
EXPECTED_DATASET="$3"

SOURCE="$JOB_DIR/source.c"
DATASET="$JOB_DIR/dataset"
PROGRAM="$JOB_DIR/program"

test -f "$SOURCE"
test -f "$DATASET"

command -v mpicc.mpich >/dev/null 2>&1

SOURCE_SHA="$(
    sha256sum "$SOURCE" |
    awk '{print $1}'
)"

DATASET_SHA="$(
    sha256sum "$DATASET" |
    awk '{print $1}'
)"

test "$SOURCE_SHA" = "$EXPECTED_SOURCE"
test "$DATASET_SHA" = "$EXPECTED_DATASET"

mpicc.mpich \
    -O2 \
    -Wall \
    -Wextra \
    -o "$PROGRAM" \
    "$SOURCE"

chmod 0644 \
    "$SOURCE" \
    "$DATASET"

chmod 0755 \
    "$PROGRAM"

test -x "$PROGRAM"

printf 'SOURCE_SHA=%s\n' "$SOURCE_SHA"
printf 'DATASET_SHA=%s\n' "$DATASET_SHA"
printf 'PROGRAM=READY\n'
`

	output, err :=
		e.runSSH(
			ctx,
			45*time.Second,
			worker,
			compileScript,
			remoteDirectory,
			sourceHash,
			datasetHash,
		)

	if err != nil {
		return err
	}

	if !strings.Contains(
		output,
		"PROGRAM=READY",
	) {

		return fmt.Errorf(
			"compilación remota no confirmó PROGRAM=READY",
		)
	}

	return nil
}

func (e *Executor) executeHPC(
	ctx context.Context,
	token string,
	jobID string,
	processes int32,
) (*Result, error) {

	commandContext,
		cancel :=
		context.WithTimeout(
			ctx,
			90*time.Second,
		)

	defer cancel()

	command :=
		exec.CommandContext(
			commandContext,
			"java",
			"-cp",
			e.javaClasspath,
			"hpc.client.ScientificJobClient",
		)

	command.Env =
		append(
			os.Environ(),
			"HPC_COORDINATOR_HOST="+
				e.coordinatorHost,
			"HPC_COORDINATOR_PORT="+
				e.coordinatorPort,
			"HPC_TOKEN="+
				token,
			"HPC_JOB_ID="+
				jobID,
			"HPC_PROCESSES="+
				strconv.Itoa(
					int(
						processes,
					),
				),
		)

	rawOutput, err :=
		command.CombinedOutput()

	if errors.Is(
		commandContext.Err(),
		context.DeadlineExceeded,
	) {

		return nil,
			fmt.Errorf(
				"timeout esperando el job HPC",
			)
	}

	if err != nil {

		return nil,
			fmt.Errorf(
				"ScientificJobClient falló: %w: %s",
				err,
				truncate(
					string(
						rawOutput,
					),
					4096,
				),
			)
	}

	return parseBridgeOutput(
		jobID,
		string(
			rawOutput,
		),
	)
}

func parseBridgeOutput(
	expectedJobID string,
	raw string,
) (*Result, error) {

	values :=
		make(
			map[string]string,
		)

	for _, line := range strings.Split(
		raw,
		"\n",
	) {

		key, value, ok :=
			strings.Cut(
				line,
				"=",
			)

		if !ok {
			continue
		}

		switch key {

		case "JOB_ID",
			"SUCCESS",
			"EXIT_CODE",
			"NODE_ID",
			"DURATION_MS",
			"OUTPUT_BASE64":

			values[key] = value
		}
	}

	if values["JOB_ID"] !=
		expectedJobID {

		return nil,
			fmt.Errorf(
				"JOB_ID inválido en respuesta HPC",
			)
	}

	success, err :=
		strconv.ParseBool(
			values["SUCCESS"],
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"SUCCESS inválido en respuesta HPC",
			)
	}

	exitCode, err :=
		strconv.ParseInt(
			values["EXIT_CODE"],
			10,
			32,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"EXIT_CODE inválido en respuesta HPC",
			)
	}

	duration, err :=
		strconv.ParseInt(
			values["DURATION_MS"],
			10,
			64,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"DURATION_MS inválido en respuesta HPC",
			)
	}

	outputBytes, err :=
		base64.StdEncoding.DecodeString(
			values["OUTPUT_BASE64"],
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"OUTPUT_BASE64 inválido: %w",
				err,
			)
	}

	return &Result{
		JobID:        expectedJobID,
		Success:      success,
		ExitCode:     int32(exitCode),
		LauncherNode: values["NODE_ID"],
		DurationMs:   duration,
		Output: string(
			outputBytes,
		),
	}, nil
}

func (e *Executor) runSSH(
	ctx context.Context,
	timeout time.Duration,
	worker string,
	script string,
	arguments ...string,
) (string, error) {

	args :=
		[]string{
			"-o",
			"BatchMode=yes",
			"-o",
			"ConnectTimeout=5",
			worker,
			"bash",
			"-s",
			"--",
		}

	args =
		append(
			args,
			arguments...,
		)

	return runCommand(
		ctx,
		timeout,
		script,
		"ssh",
		args...,
	)
}

func (e *Executor) cleanupRemote(
	jobID string,
) {

	if !isUUID(
		jobID,
	) {
		return
	}

	script := `set -eu
JOB_DIR="$1"
rm -rf -- "$JOB_DIR"
`

	for _, worker := range e.workers {

		ctx, cancel :=
			context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

		_, _ =
			e.runSSH(
				ctx,
				5*time.Second,
				worker,
				script,
				e.remoteJobRoot+
					"/"+
					jobID,
			)

		cancel()
	}
}

func runCommand(
	ctx context.Context,
	timeout time.Duration,
	stdin string,
	name string,
	arguments ...string,
) (string, error) {

	commandContext,
		cancel :=
		context.WithTimeout(
			ctx,
			timeout,
		)

	defer cancel()

	command :=
		exec.CommandContext(
			commandContext,
			name,
			arguments...,
		)

	if stdin != "" {

		command.Stdin =
			strings.NewReader(
				stdin,
			)
	}

	output, err :=
		command.CombinedOutput()

	if errors.Is(
		commandContext.Err(),
		context.DeadlineExceeded,
	) {

		return string(output),
			fmt.Errorf(
				"timeout ejecutando %s",
				name,
			)
	}

	if err != nil {

		return string(output),
			fmt.Errorf(
				"%s falló: %w: %s",
				name,
				err,
				truncate(
					string(output),
					4096,
				),
			)
	}

	return string(output), nil
}

func sha256Hex(
	value []byte,
) string {

	sum :=
		sha256.Sum256(
			value,
		)

	return hex.EncodeToString(
		sum[:],
	)
}

func newUUID() (
	string,
	error,
) {

	value :=
		make(
			[]byte,
			16,
		)

	if _, err :=
		rand.Read(
			value,
		); err != nil {

		return "",
			fmt.Errorf(
				"no se pudo generar UUID: %w",
				err,
			)
	}

	value[6] =
		(value[6] & 0x0f) |
			0x40

	value[8] =
		(value[8] & 0x3f) |
			0x80

	raw :=
		hex.EncodeToString(
			value,
		)

	return raw[0:8] +
			"-" +
			raw[8:12] +
			"-" +
			raw[12:16] +
			"-" +
			raw[16:20] +
			"-" +
			raw[20:32],
		nil
}

func isUUID(
	value string,
) bool {

	if len(value) != 36 {
		return false
	}

	for index, char := range value {

		switch index {

		case 8, 13, 18, 23:

			if char != '-' {
				return false
			}

			continue
		}

		switch {

		case char >= '0' &&
			char <= '9':

		case char >= 'a' &&
			char <= 'f':

		case char >= 'A' &&
			char <= 'F':

		default:
			return false
		}
	}

	return true
}

func getEnv(
	name string,
	fallback string,
) string {

	value :=
		strings.TrimSpace(
			os.Getenv(
				name,
			),
		)

	if value == "" {
		return fallback
	}

	return value
}

func truncate(
	value string,
	maximum int,
) string {

	if len(value) <= maximum {
		return value
	}

	return value[:maximum] +
		"..."
}
