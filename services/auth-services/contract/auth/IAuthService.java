package auth;

import java.rmi.Remote;
import java.rmi.RemoteException;

public interface IAuthService extends Remote {

    AuthResult login(
            String username,
            String password
    ) throws RemoteException;

    TokenResult validateToken(
            String token
    ) throws RemoteException;

    AuthResult getUser(
            String userId
    ) throws RemoteException;
}