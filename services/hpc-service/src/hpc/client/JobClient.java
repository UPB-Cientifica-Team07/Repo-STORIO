package hpc.client;

import hpc.common.ClusterCoordinator;
import hpc.common.MpiJobResult;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.util.UUID;

public class JobClient {

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

            if (
                args.length < 3
            ) {

                throw new IllegalArgumentException(
                    "Uso: JobClient <host> <port> <token> [processes]"
                );
            }

            String token =
                args[2];

            int processes =
                args.length > 3
                    ? Integer.parseInt(
                        args[3]
                    )
                    : 4;

            String programName =
                args.length > 4
                    ? args[4]
                    : "hello_mpi";

            Registry registry =
                LocateRegistry
                    .getRegistry(
                        host,
                        port
                    );

            ClusterCoordinator coordinator =
                (ClusterCoordinator)
                    registry.lookup(
                        ClusterCoordinator
                            .SERVICE_NAME
                    );

            String jobId =
                UUID
                    .randomUUID()
                    .toString();

            System.out.println(
                "Enviando job: " +
                jobId
            );

            MpiJobResult result =
                coordinator.submitMpiJob(
                    token,
                    jobId,
                    programName,
                    processes
                );

            System.out.println(
                "==================================="
            );
            System.out.println(
                " RESULTADO HPC"
            );
            System.out.println(
                " Job: " +
                result.getJobId()
            );
            System.out.println(
                " Nodo: " +
                result.getNodeId()
            );
            System.out.println(
                " Success: " +
                result.isSuccess()
            );
            System.out.println(
                " Exit code: " +
                result.getExitCode()
            );
            System.out.println(
                " Duración: " +
                result.getDurationMs() +
                " ms"
            );
            System.out.println(
                "-----------------------------------"
            );
            System.out.print(
                result.getOutput()
            );
            System.out.println(
                "==================================="
            );

        } catch (Exception error) {

            error.printStackTrace();

            System.exit(1);
        }
    }
}
