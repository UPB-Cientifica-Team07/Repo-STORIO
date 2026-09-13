package hpc.client;

import hpc.common.ClusterCoordinator;
import hpc.common.NodeInfo;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.util.List;

public final class HealthClient {

    private HealthClient() {
    }

    public static void main(
        String[] args
    ) {

        try {

            String host =
                args.length > 0
                    ? args[0]
                    : "127.0.0.1";

            int port =
                args.length > 1
                    ? Integer.parseInt(
                        args[1]
                    )
                    : 1100;

            Registry registry =
                LocateRegistry.getRegistry(
                    host,
                    port
                );

            ClusterCoordinator coordinator =
                (ClusterCoordinator)
                    registry.lookup(
                        ClusterCoordinator.SERVICE_NAME
                    );

            System.out.println(
                "health=" +
                coordinator.health()
            );

            List<NodeInfo> nodes =
                coordinator.listNodes();

            System.out.println(
                "nodes=" +
                nodes.size()
            );

            for (NodeInfo node : nodes) {

                System.out.println(
                    "node_id=" +
                    node.getNodeId() +
                    "|hostname=" +
                    node.getHostname() +
                    "|ip=" +
                    node.getIp() +
                    "|cpu=" +
                    node.getCpuCores() +
                    "|memory_mb=" +
                    node.getMemoryMb() +
                    "|status=" +
                    node.getStatus()
                );
            }

        } catch (Exception error) {

            error.printStackTrace();

            System.exit(1);
        }
    }
}
