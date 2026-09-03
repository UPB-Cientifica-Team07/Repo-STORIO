package hpc.coordinator;

import hpc.common.ClusterCoordinator;

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

            Registry registry =
                LocateRegistry
                    .createRegistry(
                        port
                    );

            ClusterCoordinator coordinator =
                new ClusterCoordinatorImpl();

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
                " Puerto RMI: " + port
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
