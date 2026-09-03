package hpc.common;

import java.io.Serializable;

public class NodeInfo implements Serializable {

    private static final long serialVersionUID = 1L;

    private final String nodeId;
    private final String hostname;
    private final String ip;
    private final int cpuCores;
    private final long memoryMb;
    private final String status;

    public NodeInfo(
        String nodeId,
        String hostname,
        String ip,
        int cpuCores,
        long memoryMb,
        String status
    ) {
        this.nodeId = nodeId;
        this.hostname = hostname;
        this.ip = ip;
        this.cpuCores = cpuCores;
        this.memoryMb = memoryMb;
        this.status = status;
    }

    public String getNodeId() {
        return nodeId;
    }

    public String getHostname() {
        return hostname;
    }

    public String getIp() {
        return ip;
    }

    public int getCpuCores() {
        return cpuCores;
    }

    public long getMemoryMb() {
        return memoryMb;
    }

    public String getStatus() {
        return status;
    }

    @Override
    public String toString() {
        return String.format(
            "%s | %s | %s | CPU=%d | RAM=%d MB | %s",
            nodeId,
            hostname,
            ip,
            cpuCores,
            memoryMb,
            status
        );
    }
}
