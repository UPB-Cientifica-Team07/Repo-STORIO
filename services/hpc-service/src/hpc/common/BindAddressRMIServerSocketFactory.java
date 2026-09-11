package hpc.common;

import java.io.IOException;
import java.io.Serializable;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.ServerSocket;

import java.rmi.server.RMIServerSocketFactory;

import java.util.Objects;

public final class BindAddressRMIServerSocketFactory
    implements RMIServerSocketFactory, Serializable {

    private static final long serialVersionUID =
        1L;

    private final String bindAddress;

    public BindAddressRMIServerSocketFactory(
        String bindAddress
    ) {

        if (
            bindAddress == null ||
            bindAddress.isBlank()
        ) {
            throw new IllegalArgumentException(
                "La dirección RMI no puede estar vacía"
            );
        }

        this.bindAddress =
            bindAddress;
    }

    @Override
    public ServerSocket createServerSocket(
        int port
    ) throws IOException {

        ServerSocket socket =
            new ServerSocket();

        socket.bind(
            new InetSocketAddress(
                InetAddress.getByName(
                    bindAddress
                ),
                port
            )
        );

        return socket;
    }

    @Override
    public boolean equals(
        Object other
    ) {

        if (this == other) {
            return true;
        }

        if (
            !(
                other instanceof
                    BindAddressRMIServerSocketFactory
            )
        ) {
            return false;
        }

        BindAddressRMIServerSocketFactory that =
            (BindAddressRMIServerSocketFactory)
                other;

        return bindAddress.equals(
            that.bindAddress
        );
    }

    @Override
    public int hashCode() {

        return Objects.hash(
            bindAddress
        );
    }
}
