package hpc.common;

import java.rmi.Remote;
import java.rmi.RemoteException;
import java.util.List;

public interface ClusterCoordinator
    extends Remote {

    String SERVICE_NAME =
        "UPB_HPC_COORDINATOR";

    boolean registerNode(
        NodeInfo node,
        HpcWorker worker
    ) throws RemoteException;

    boolean heartbeat(
        String nodeId
    ) throws RemoteException;

    List<NodeInfo> listNodes()
        throws RemoteException;

    MpiJobResult submitMpiJob(
        String token,
        String jobId,
        String programName,
        int processes
    ) throws RemoteException;

    String health()
        throws RemoteException;
}
