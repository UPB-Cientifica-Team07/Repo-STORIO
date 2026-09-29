package hpc.scheduler;

import hpc.common.NodeInfo;

import java.time.Duration;
import java.time.Instant;
import java.util.Map;
import java.util.Set;

public class HpcScheduler {

    private final Map<String, NodeInfo>
        nodes;

    private final Map<String, Instant>
        heartbeats;

    private final Set<String>
        busyNodes;

    private final Duration
        heartbeatTimeout;

    public HpcScheduler(
        Map<String, NodeInfo> nodes,
        Map<String, Instant> heartbeats,
        Set<String> busyNodes,
        Duration heartbeatTimeout
    ) {

        this.nodes =
            nodes;

        this.heartbeats =
            heartbeats;

        this.busyNodes =
            busyNodes;

        this.heartbeatTimeout =
            heartbeatTimeout;
    }

    public synchronized String selectAndReserveNode(
        int processes
    ) {

        if (
            processes <= 0
        ) {
            return null;
        }

        /*
         * El hostfile MPI representa el cluster completo.
         *
         * WorkerNode actúa como launcher de MPICH, por lo que
         * la capacidad requerida por el job debe compararse
         * contra la suma de CPU de los nodos vivos y no contra
         * la CPU del launcher individual.
         *
         * El cluster usa un hostfile compartido. Mientras exista
         * un job reservado no permitimos otro lanzamiento MPI
         * concurrente sobre el mismo conjunto de nodos.
         */
        if (
            !busyNodes.isEmpty()
        ) {
            return null;
        }

        Instant now =
            Instant.now();

        String bestNodeId =
            null;

        int bestCpu =
            Integer.MAX_VALUE;

        int availableCpu =
            0;

        for (
            Map.Entry<String, NodeInfo> entry
                : nodes.entrySet()
        ) {

            String nodeId =
                entry.getKey();

            NodeInfo node =
                entry.getValue();

            Instant lastHeartbeat =
                heartbeats.get(
                    nodeId
                );

            if (
                lastHeartbeat == null
            ) {
                continue;
            }

            Duration age =
                Duration.between(
                    lastHeartbeat,
                    now
                );

            if (
                age.compareTo(
                    heartbeatTimeout
                ) > 0
            ) {
                continue;
            }

            if (
                busyNodes.contains(
                    nodeId
                )
            ) {
                continue;
            }

            availableCpu +=
                node.getCpuCores();

            /*
             * Best-fit:
             * selecciona el nodo suficiente
             * con menor cantidad de CPU.
             *
             * Así evitamos desperdiciar
             * un nodo grande si uno menor
             * puede ejecutar el trabajo.
             */
            if (
                node.getCpuCores()
                    < bestCpu
            ) {

                bestCpu =
                    node.getCpuCores();

                bestNodeId =
                    nodeId;
            }
        }

        if (
            bestNodeId != null &&
            availableCpu >= processes
        ) {

            busyNodes.add(
                bestNodeId
            );

        } else {

            bestNodeId =
                null;
        }

        return bestNodeId;
    }

    public void releaseNode(
        String nodeId
    ) {

        if (
            nodeId != null
        ) {

            busyNodes.remove(
                nodeId
            );
        }
    }

    public boolean isAlive(
        String nodeId
    ) {

        Instant lastHeartbeat =
            heartbeats.get(
                nodeId
            );

        if (
            lastHeartbeat == null
        ) {
            return false;
        }

        Duration age =
            Duration.between(
                lastHeartbeat,
                Instant.now()
            );

        return age.compareTo(
            heartbeatTimeout
        ) <= 0;
    }

    public boolean isBusy(
        String nodeId
    ) {

        return busyNodes.contains(
            nodeId
        );
    }
}
