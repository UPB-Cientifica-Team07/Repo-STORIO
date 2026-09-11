package hpc.coordinator;

import hpc.common.ClusterCoordinator;
import hpc.common.BindAddressRMIServerSocketFactory;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class CoordinatorServer {

    private static final int DEFAULT_PORT =
        1100;

    public static void main(
        String[] args
    ) {

        try {

            int port =
                Integer.parseInt(
                    System.getenv()
                        .getOrDefault(
                            "HPC_RMI_PORT",
                            String.valueOf(
                                DEFAULT_PORT
                            )
                        )
                );

            String bindAddress =
                System.getenv()
                    .getOrDefault(
                        "HPC_BIND_ADDRESS",
                        "127.0.0.1"
                    );

            String advertiseAddress =
                System.getenv()
                    .getOrDefault(
                        "HPC_ADVERTISE_ADDRESS",
                        bindAddress
                    );

            int exportPort =
                Integer.parseInt(
                    System.getenv()
                        .getOrDefault(
                            "HPC_COORDINATOR_RMI_PORT",
                            "1102"
                        )
                );

            System.setProperty(
                "java.rmi.server.hostname",
                advertiseAddress
            );

            BindAddressRMIServerSocketFactory
                serverSocketFactory =
                    new BindAddressRMIServerSocketFactory(
                        bindAddress
                    );

            Registry registry =
                LocateRegistry
                    .createRegistry(
                        port,
                        null,
                        serverSocketFactory
                    );

            ClusterCoordinator coordinator =
                new ClusterCoordinatorImpl(
                    exportPort,
                    serverSocketFactory
                );

            registry.rebind(
                ClusterCoordinator.SERVICE_NAME,
                coordinator
            );

            System.out.println(
                "==================================="
            );
            System.out.println(
                " HPC COORDINATOR"
            );
            System.out.println(
                " Tecnología: Java RMI"
            );
            System.out.println(
                " Registry RMI: " +
                bindAddress +
                ":" +
                port
            );
            System.out.println(
                " Objeto Coordinator: " +
                bindAddress +
                ":" +
                exportPort
            );
            System.out.println(
                " Dirección anunciada: " +
                advertiseAddress
            );
            System.out.println(
                " Servicio: " +
                ClusterCoordinator.SERVICE_NAME
            );
            System.out.println(
                " Estado: ACTIVE"
            );
            System.out.println(
                "==================================="
            );

            Thread.currentThread()
                .join();

        } catch (Exception error) {

            System.err.println(
                "No se pudo iniciar HPC Coordinator:"
            );

            error.printStackTrace();

            System.exit(1);
        }
    }
}
