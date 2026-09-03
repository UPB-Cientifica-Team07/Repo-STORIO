package hpc.common;

import java.rmi.Remote;
import java.rmi.RemoteException;

public interface HpcWorker
    extends Remote {

    MpiJobResult executeMpi(
        String jobId,
        String programName,
        int processes
    ) throws RemoteException;
}
