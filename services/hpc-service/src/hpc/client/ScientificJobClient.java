package hpc.client;

import hpc.common.ClusterCoordinator;
import hpc.common.MpiJobResult;

import java.nio.charset.StandardCharsets;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

import java.util.Base64;

public final class ScientificJobClient {

    private ScientificJobClient() {
    }

    public static void main(
        String[] args
    ) {

        try {

            String host =
                requireEnv(
                    "HPC_COORDINATOR_HOST"
                );

            int port =
                Integer.parseInt(
                    requireEnv(
                        "HPC_COORDINATOR_PORT"
                    )
                );

            String token =
                requireEnv(
                    "HPC_TOKEN"
                );

            String jobId =
                requireEnv(
                    "HPC_JOB_ID"
                );

            int processes =
                Integer.parseInt(
                    requireEnv(
                        "HPC_PROCESSES"
                    )
                );

            if (processes < 1) {

                throw new IllegalArgumentException(
                    "HPC_PROCESSES debe ser mayor que cero"
                );
            }

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

            MpiJobResult result =
                coordinator.submitMpiJob(
                    token,
                    jobId,
                    "scientific_job",
                    processes
                );

            String output =
                result.getOutput() == null
                    ? ""
                    : result.getOutput();

            String outputBase64 =
                Base64
                    .getEncoder()
                    .encodeToString(
                        output.getBytes(
                            StandardCharsets.UTF_8
                        )
                    );

            System.out.println(
                "JOB_ID=" +
                result.getJobId()
            );

            System.out.println(
                "SUCCESS=" +
                result.isSuccess()
            );

            System.out.println(
                "EXIT_CODE=" +
                result.getExitCode()
            );

            System.out.println(
                "NODE_ID=" +
                result.getNodeId()
            );

            System.out.println(
                "DURATION_MS=" +
                result.getDurationMs()
            );

            System.out.println(
                "OUTPUT_BASE64=" +
                outputBase64
            );

        } catch (Exception error) {

            System.err.println(
                "SCIENTIFIC_CLIENT_ERROR=" +
                error.getMessage()
            );

            error.printStackTrace(
                System.err
            );

            System.exit(1);
        }
    }

    private static String requireEnv(
        String name
    ) {

        String value =
            System.getenv(
                name
            );

        if (
            value == null ||
            value.isBlank()
        ) {

            throw new IllegalStateException(
                name +
                " es obligatorio"
            );
        }

        return value.trim();
    }
}
