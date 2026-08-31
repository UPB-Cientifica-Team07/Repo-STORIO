package co.edu.upb.cientifica.sync.network

import android.content.Context
import android.net.ConnectivityManager
import android.net.LinkProperties
import android.net.Network
import android.net.NetworkCapabilities

class NetworkMonitor(
    context: Context,
    private val onNetworkChanged: () -> Unit
) {

    private val connectivityManager =
        context.getSystemService(
            Context.CONNECTIVITY_SERVICE
        ) as ConnectivityManager

    private val callback =
        object :
            ConnectivityManager.NetworkCallback() {

            override fun onAvailable(
                network: Network
            ) {

                onNetworkChanged()
            }

            override fun onLost(
                network: Network
            ) {

                onNetworkChanged()
            }

            override fun onCapabilitiesChanged(
                network: Network,
                networkCapabilities: NetworkCapabilities
            ) {

                onNetworkChanged()
            }

            override fun onLinkPropertiesChanged(
                network: Network,
                linkProperties: LinkProperties
            ) {

                onNetworkChanged()
            }
        }

    fun start() {

        connectivityManager
            .registerDefaultNetworkCallback(
                callback
            )
    }

    fun stop() {

        try {

            connectivityManager
                .unregisterNetworkCallback(
                    callback
                )

        } catch (_: Exception) {
        }
    }
}
