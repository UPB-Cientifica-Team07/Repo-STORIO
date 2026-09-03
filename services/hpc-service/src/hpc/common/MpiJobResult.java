package hpc.common;

import java.io.Serializable;

public class MpiJobResult
    implements Serializable {

    private static final long serialVersionUID =
        1L;

    private final String jobId;
    private final String nodeId;
    private final boolean success;
    private final int exitCode;
    private final String output;
    private final long durationMs;

    public MpiJobResult(
        String jobId,
        String nodeId,
        boolean success,
        int exitCode,
        String output,
        long durationMs
    ) {
        this.jobId = jobId;
        this.nodeId = nodeId;
        this.success = success;
        this.exitCode = exitCode;
        this.output = output;
        this.durationMs = durationMs;
    }

    public String getJobId() {
        return jobId;
    }

    public String getNodeId() {
        return nodeId;
    }

    public boolean isSuccess() {
        return success;
    }

    public int getExitCode() {
        return exitCode;
    }

    public String getOutput() {
        return output;
    }

    public long getDurationMs() {
        return durationMs;
    }
}
