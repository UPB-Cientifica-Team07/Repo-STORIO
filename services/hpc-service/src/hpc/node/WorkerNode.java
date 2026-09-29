package hpc.node;

import hpc.common.ClusterCoordinator;
import hpc.common.BindAddressRMIServerSocketFactory;
import hpc.common.HpcWorker;
import hpc.common.MpiJobResult;
import hpc.common.NodeInfo;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.InetAddress;
import java.nio.file.Path;
import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.rmi.server.UnicastRemoteObject;
import java.rmi.server.RMIServerSocketFactory;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

public class WorkerNode
    extends UnicastRemoteObject
    implements HpcWorker {

    private static final Set<String>
        ALLOWED_PROGRAMS =
            Set.of(
                "hello_mpi",
                "sleep_mpi"
            );

    private final String nodeId;

    private final Path mpiDirectory;

    protected WorkerNode(
        String nodeId,
        Path mpiDirectory,
        int exportPort,
        RMIServerSocketFactory serverSocketFactory
    ) throws Exception {

        super(
            exportPort,
            null,
            serverSocketFactory
        );

        this.nodeId =
            nodeId;

        this.mpiDirectory =
            mpiDirectory;
    }

    @Override
    public MpiJobResult executeMpi(
        String jobId,
        String programName,
        int processes
    ) {

        long start =
            System.currentTimeMillis();

        if (
            !ALLOWED_PROGRAMS.contains(
                programName
            )
        ) {

            return new MpiJobResult(
                jobId,
                nodeId,
                false,
                -1,
                "Programa MPI no permitido",
                0
            );
        }

        int maxProcesses =
            Runtime
                .getRuntime()
                .availableProcessors();

        String maxProcessesOverride =
            System.getenv(
                "HPC_MPI_MAX_PROCESSES"
            );

        if (
            maxProcessesOverride != null &&
            !maxProcessesOverride.isBlank()
        ) {

            try {

                maxProcesses =
                    Integer.parseInt(
                        maxProcessesOverride
                    );

            } catch (
                NumberFormatException error
            ) {

                return new MpiJobResult(
                    jobId,
                    nodeId,
                    false,
                    -1,
                    "HPC_MPI_MAX_PROCESSES inválido",
                    0
                );
            }
        }

        if (
            processes < 1 ||
            processes > maxProcesses
        ) {

            return new MpiJobResult(
                jobId,
                nodeId,
                false,
                -1,
                "Cantidad de procesos inválida. Máximo: " +
                    maxProcesses,
                0
            );
        }

        Path executable =
            mpiDirectory
                .resolve(
                    programName
                )
                .normalize();

        try {

            if (
                !executable
                    .toFile()
                    .isFile()
            ) {

                throw new Exception(
                    "Ejecutable MPI no encontrado: " +
                    executable
                );
            }

            java.util.List<String> mpiCommand =
                new java.util.ArrayList<>();

            String mpiLauncher =
                System.getenv()
                    .getOrDefault(
                        "HPC_MPI_LAUNCHER",
                        "mpirun"
                    )
                    .trim();

            String mpiFlavor =
                System.getenv()
                    .getOrDefault(
                        "HPC_MPI_FLAVOR",
                        "openmpi"
                    )
                    .trim();

            String mpiInterface =
                System.getenv(
                    "HPC_MPI_INTERFACE"
                );

            if (mpiLauncher.isBlank()) {

                throw new Exception(
                    "HPC_MPI_LAUNCHER vacío"
                );
            }

            mpiCommand.add(
                mpiLauncher
            );

            String hostfile =
                System.getenv(
                    "HPC_MPI_HOSTFILE"
                );

            Path hostfilePath = null;

            if (
                hostfile != null &&
                !hostfile.isBlank()
            ) {

                hostfilePath =
                    Path.of(
                        hostfile
                    )
                    .toAbsolutePath()
                    .normalize();

                if (
                    !hostfilePath
                        .toFile()
                        .isFile()
                ) {

                    throw new Exception(
                        "Hostfile MPI no encontrado: " +
                        hostfilePath
                    );
                }
            }

            if (
                "mpich".equalsIgnoreCase(
                    mpiFlavor
                )
            ) {

                mpiCommand.add(
                    "-launcher"
                );

                mpiCommand.add(
                    "ssh"
                );

                if (
                    mpiInterface != null &&
                    !mpiInterface.isBlank()
                ) {

                    mpiCommand.add(
                        "-iface"
                    );

                    mpiCommand.add(
                        mpiInterface.trim()
                    );
                }

                if (hostfilePath != null) {

                    mpiCommand.add(
                        "-f"
                    );

                    mpiCommand.add(
                        hostfilePath
                            .toString()
                    );
                }

            } else if (
                "openmpi".equalsIgnoreCase(
                    mpiFlavor
                )
            ) {

                if (hostfilePath != null) {

                    mpiCommand.add(
                        "--hostfile"
                    );

                    mpiCommand.add(
                        hostfilePath
                            .toString()
                    );

                    mpiCommand.add(
                        "--map-by"
                    );

                    mpiCommand.add(
                        "slot"
                    );
                }

            } else {

                throw new Exception(
                    "HPC_MPI_FLAVOR no soportado: " +
                    mpiFlavor
                );
            }

            mpiCommand.add(
                "-np"
            );

            mpiCommand.add(
                String.valueOf(
                    processes
                )
            );

            mpiCommand.add(
                executable
                    .toAbsolutePath()
                    .toString()
            );

            System.out.println(
                "[MPI] Comando: " +
                String.join(
                    " ",
                    mpiCommand
                )
            );

            ProcessBuilder builder =
                new ProcessBuilder(
                    mpiCommand
                );

            builder
                .redirectErrorStream(
                    true
                );

            System.out.println(
                "[MPI] Ejecutando job " +
                jobId +
                " con " +
                processes +
                " procesos"
            );

            Process process =
                builder.start();

            StringBuilder output =
                new StringBuilder();

            try (
                BufferedReader reader =
                    new BufferedReader(
                        new InputStreamReader(
                            process.getInputStream()
                        )
                    )
            ) {

                String line;

                while (
                    (
                        line =
                            reader.readLine()
                    ) != null
                ) {

                    output
                        .append(line)
                        .append(
                            System.lineSeparator()
                        );
                }
            }

            boolean finished =
                process.waitFor(
                    60,
                    TimeUnit.SECONDS
                );

            if (!finished) {

                process.destroyForcibly();

                return new MpiJobResult(
                    jobId,
                    nodeId,
                    false,
                    -1,
                    "Timeout ejecutando MPI",
                    System.currentTimeMillis()
                        - start
                );
            }

            int exitCode =
                process.exitValue();

            return new MpiJobResult(
                jobId,
                nodeId,
                exitCode == 0,
                exitCode,
                output.toString(),
                System.currentTimeMillis()
                    - start
            );

        } catch (Exception error) {

            return new MpiJobResult(
                jobId,
                nodeId,
                false,
                -1,
                error.getMessage(),
                System.currentTimeMillis()
                    - start
            );
        }
    }

    public static void main(
        String[] args
    ) {

        try {

            String coordinatorHost =
                args.length > 0
                    ? args[0]
                    : "127.0.0.1";

            int coordinatorPort =
                args.length > 1
                    ? Integer.parseInt(
                        args[1]
                    )
                    : 1100;

            String nodeId =
                args.length > 2
                    ? args[2]
                    : UUID
                        .randomUUID()
                        .toString();

            String logicalHostname =
                args.length > 3
                    ? args[3]
                    : nodeId;

            Integer cpuOverride =
                args.length > 4
                    ? Integer.valueOf(
                        args[4]
                    )
                    : null;

            Path mpiDirectory =
                Path.of(
                    System.getenv()
                        .getOrDefault(
                            "HPC_MPI_DIR",
                            "services/hpc-service/mpi"
                        )
                )
                .toAbsolutePath()
                .normalize();

            InetAddress localHost =
                InetAddress
                    .getLocalHost();

            String physicalHostname =
                localHost
                    .getHostName();

            String hostname =
                logicalHostname;

            String ip =
                localHost
                    .getHostAddress();

            int detectedCpu =
                Runtime
                    .getRuntime()
                    .availableProcessors();

            int cpuCores =
                cpuOverride != null
                    ? cpuOverride
                    : detectedCpu;

            if (
                cpuCores < 1 ||
                cpuCores > detectedCpu
            ) {

                throw new IllegalArgumentException(
                    "CPU lógica inválida. Debe estar entre 1 y " +
                    detectedCpu
                );
            }

            long memoryMb =
                Runtime
                    .getRuntime()
                    .maxMemory()
                    /
                    (1024L * 1024L);

            if (
                args.length < 6
            ) {
                throw new IllegalArgumentException(
                    "Uso: WorkerNode <coordinatorHost> "
                    + "<coordinatorPort> <nodeId> "
                    + "<logicalHostname> <cpuCores> "
                    + "<workerRmiPort>"
                );
            }

            int workerRmiPort =
                Integer.parseInt(
                    args[5]
                );

            String workerBindAddress =
                System.getenv()
                    .getOrDefault(
                        "HPC_WORKER_BIND_ADDRESS",
                        "127.0.0.1"
                    );

            String workerAdvertiseAddress =
                System.getenv()
                    .getOrDefault(
                        "HPC_WORKER_ADVERTISE_ADDRESS",
                        workerBindAddress
                    );

            System.setProperty(
                "java.rmi.server.hostname",
                workerAdvertiseAddress
            );

            BindAddressRMIServerSocketFactory
                workerSocketFactory =
                    new BindAddressRMIServerSocketFactory(
                        workerBindAddress
                    );

            WorkerNode worker =
                new WorkerNode(
                    nodeId,
                    mpiDirectory,
                    workerRmiPort,
                    workerSocketFactory
                );

            NodeInfo node =
                new NodeInfo(
                    nodeId,
                    hostname,
                    ip,
                    cpuCores,
                    memoryMb,
                    "DISPONIBLE"
                );

            Registry registry =
                LocateRegistry
                    .getRegistry(
                        coordinatorHost,
                        coordinatorPort
                    );

            ClusterCoordinator coordinator =
                (ClusterCoordinator)
                    registry.lookup(
                        ClusterCoordinator
                            .SERVICE_NAME
                    );

            boolean registered =
                coordinator.registerNode(
                    node,
                    worker
                );

            if (!registered) {

                throw new IllegalStateException(
                    "El coordinador rechazó el nodo"
                );
            }

            System.out.println(
                "==================================="
            );
            System.out.println(
                " HPC WORKER NODE"
            );
            System.out.println(
                " Nodo: " + nodeId
            );
            System.out.println(
                " Hostname lógico: " +
                hostname
            );
            System.out.println(
                " Host físico: " +
                physicalHostname
            );
            System.out.println(
                " IP: " + ip
            );
            System.out.println(
                " CPU: " + cpuCores
            );
            System.out.println(
                " MPI DIR: " + mpiDirectory
            );
            System.out.println(
                " Estado: REGISTERED"
            );
            System.out.println(
                "==================================="
            );

            while (true) {

                Thread.sleep(
                    5000
                );

                coordinator.heartbeat(
                    nodeId
                );
            }

        } catch (Exception error) {

            System.err.println(
                "Error en HPC Worker:"
            );

            error.printStackTrace();

            System.exit(1);
        }
    }
}
