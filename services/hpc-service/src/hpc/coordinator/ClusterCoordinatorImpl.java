package hpc.coordinator;

import hpc.common.ClusterCoordinator;
import hpc.common.HpcWorker;
import hpc.common.MpiJobResult;
import hpc.common.NodeInfo;
import hpc.repository.JobRepository;
import hpc.repository.NodeRepository;
import hpc.repository.UserRepository;
import hpc.auth.AuthClient;
import hpc.auth.AuthIdentity;
import hpc.scheduler.HpcScheduler;

import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Map;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class ClusterCoordinatorImpl
    extends UnicastRemoteObject
    implements ClusterCoordinator {

    private final Map<String, NodeInfo>
        nodes =
            new ConcurrentHashMap<>();

    private final Map<String, HpcWorker>
        workers =
            new ConcurrentHashMap<>();

    private final Map<String, Instant>
        heartbeats =
            new ConcurrentHashMap<>();

    private final Map<String, UUID>
        databaseNodeIds =
            new ConcurrentHashMap<>();

    private final Set<String>
        busyNodes =
            ConcurrentHashMap.newKeySet();

    private static final Duration
        HEARTBEAT_TIMEOUT =
            Duration.ofSeconds(
                15
            );

    private final HpcScheduler
        scheduler =
            new HpcScheduler(
                nodes,
                heartbeats,
                busyNodes,
                HEARTBEAT_TIMEOUT
            );

    private final ScheduledExecutorService
        nodeSupervisor =
            Executors
                .newSingleThreadScheduledExecutor();

    private final NodeRepository
        nodeRepository =
            new NodeRepository();

    private final JobRepository
        jobRepository =
            new JobRepository();

    private final UserRepository
        userRepository =
            new UserRepository();

    private final AuthClient
        authClient =
            new AuthClient(
                System.getenv()
                    .getOrDefault(
                        "HPC_AUTH_SERVICE",
                        "http://127.0.0.1:8081"
                    )
            );

    public ClusterCoordinatorImpl()
        throws RemoteException {

        super();

        nodeSupervisor
            .scheduleAtFixedRate(
                this::checkInactiveNodes,
                5,
                5,
                TimeUnit.SECONDS
            );
    }

    private void checkInactiveNodes() {

        for (
            String nodeId
                : nodes.keySet()
        ) {

            if (
                scheduler.isAlive(
                    nodeId
                )
            ) {
                continue;
            }

            UUID databaseNodeId =
                databaseNodeIds.get(
                    nodeId
                );

            if (
                databaseNodeId == null
            ) {
                continue;
            }

            try {

                nodeRepository
                    .updateStatus(
                        databaseNodeId,
                        "INACTIVO"
                    );

                busyNodes.remove(
                    nodeId
                );

                System.out.println(
                    "[SCHEDULER] Nodo INACTIVO: " +
                    nodeId
                );

            } catch (Exception error) {

                System.err.println(
                    "[SCHEDULER] No se pudo marcar " +
                    nodeId +
                    " como INACTIVO: " +
                    error.getMessage()
                );
            }
        }
    }

    @Override
    public boolean registerNode(
        NodeInfo node,
        HpcWorker worker
    ) throws RemoteException {

        if (
            node == null ||
            worker == null ||
            node.getNodeId() == null ||
            node.getNodeId().isBlank()
        ) {
            return false;
        }

        try {

            UUID databaseNodeId =
                nodeRepository
                    .upsertNode(
                        node
                    );

            nodes.put(
                node.getNodeId(),
                node
            );

            workers.put(
                node.getNodeId(),
                worker
            );

            databaseNodeIds.put(
                node.getNodeId(),
                databaseNodeId
            );

            heartbeats.put(
                node.getNodeId(),
                Instant.now()
            );

            System.out.println(
                "[RMI] Nodo registrado: " +
                node
            );

            System.out.println(
                "[DB] nodo_hpc id: " +
                databaseNodeId
            );

            return true;

        } catch (Exception error) {

            System.err.println(
                "[DB] Error registrando nodo: " +
                error.getMessage()
            );

            throw new RemoteException(
                "No fue posible persistir el nodo HPC",
                error
            );
        }
    }

    @Override
    public boolean heartbeat(
        String nodeId
    ) throws RemoteException {

        if (
            nodeId == null ||
            !nodes.containsKey(
                nodeId
            )
        ) {
            return false;
        }

        heartbeats.put(
            nodeId,
            Instant.now()
        );

        UUID databaseNodeId =
            databaseNodeIds.get(
                nodeId
            );

        if (
            databaseNodeId != null &&
            !busyNodes.contains(
                nodeId
            )
        ) {

            try {

                nodeRepository
                    .updateStatus(
                        databaseNodeId,
                        "DISPONIBLE"
                    );

            } catch (Exception error) {

                System.err.println(
                    "[HEARTBEAT] No se pudo actualizar nodo: " +
                    error.getMessage()
                );
            }
        }

        return true;
    }

    @Override
    public ArrayList<NodeInfo>
    listNodes()
        throws RemoteException {

        return new ArrayList<>(
            nodes.values()
        );
    }

    @Override
    public MpiJobResult submitMpiJob(
        String token,
        String jobId,
        String programName,
        int processes
    ) throws RemoteException {

        if (workers.isEmpty()) {

            throw new RemoteException(
                "No hay nodos HPC disponibles"
            );
        }

        UUID jobUuid;

        try {

            jobUuid =
                UUID.fromString(
                    jobId
                );

        } catch (
            IllegalArgumentException error
        ) {

            throw new RemoteException(
                "jobId inválido",
                error
            );
        }

        String selectedNode =
            scheduler
                .selectAndReserveNode(
                    processes
                );

        if (
            selectedNode == null
        ) {

            throw new RemoteException(
                "No hay nodo HPC disponible con capacidad para " +
                processes +
                " procesos"
            );
        }

        HpcWorker worker =
            workers.get(
                selectedNode
            );

        if (
            worker == null
        ) {

            scheduler.releaseNode(
                selectedNode
            );

            throw new RemoteException(
                "Worker RMI no disponible: " +
                selectedNode
            );
        }

        UUID databaseNodeId =
            databaseNodeIds.get(
                selectedNode
            );

        if (databaseNodeId == null) {

            throw new RemoteException(
                "El nodo seleccionado no está persistido"
            );
        }

        final UUID userId;

        try {

            AuthIdentity identity =
                authClient.validate(
                    token
                );

            if (
                !identity.valid() ||
                identity.userId().isBlank()
            ) {

                throw new RemoteException(
                    "Token HPC inválido"
                );
            }

            userId =
                userRepository
                    .findIdByDirectoryId(
                        identity.userId()
                    );

            if (userId == null) {

                throw new RemoteException(
                    "Usuario autenticado no existe en PostgreSQL"
                );
            }

            System.out.println(
                "[AUTH] Job autenticado: " +
                identity.userId() +
                " | " +
                identity.role()
            );

        } catch (RemoteException error) {

            throw error;

        } catch (Exception error) {

            throw new RemoteException(
                "No fue posible autenticar el job HPC",
                error
            );
        }

        try {

            jobRepository
                .createJob(
                    jobUuid,
                    userId,
                    "C_MPI",
                    "processes=" +
                        processes,
                    "Ejecución MPI: " +
                        programName
                );

            jobRepository
                .markRunning(
                    jobUuid,
                    databaseNodeId
                );

            nodeRepository
                .updateStatus(
                    databaseNodeId,
                    "OCUPADO"
                );

            System.out.println(
                "[HPC] Job recibido: " +
                jobId
            );

            System.out.println(
                "[HPC] Nodo asignado: " +
                selectedNode
            );

            MpiJobResult result =
                worker.executeMpi(
                    jobId,
                    programName,
                    processes
                );

            jobRepository
                .finishJob(
                    jobUuid,
                    result.isSuccess()
                );

            nodeRepository
                .updateStatus(
                    databaseNodeId,
                    "DISPONIBLE"
                );

            scheduler.releaseNode(
                selectedNode
            );

            System.out.println(
                "[SCHEDULER] Nodo liberado: " +
                selectedNode
            );

            System.out.println(
                "[DB] Job actualizado: " +
                (
                    result.isSuccess()
                        ? "FINALIZADO"
                        : "ERROR"
                )
            );

            return result;

        } catch (Exception error) {

            try {

                jobRepository
                    .finishJob(
                        jobUuid,
                        false
                    );

                nodeRepository
                    .updateStatus(
                        databaseNodeId,
                        "DISPONIBLE"
                    );

            } catch (Exception ignored) {
            }

            scheduler.releaseNode(
                selectedNode
            );

            throw new RemoteException(
                "Error ejecutando job HPC",
                error
            );
        }
    }

    @Override
    public String health()
        throws RemoteException {

        return String.format(
            "ACTIVE|nodes=%d|timestamp=%s",
            nodes.size(),
            Instant.now()
        );
    }
}
