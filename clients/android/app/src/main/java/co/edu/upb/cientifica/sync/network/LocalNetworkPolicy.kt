package co.edu.upb.cientifica.sync.network

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import java.net.Inet4Address
import java.net.InetSocketAddress
import java.net.Socket

data class NetworkPolicyResult(
    val allowed: Boolean,
    val wifi: Boolean,
    val localIp: String?,
    val subnetOk: Boolean,
    val serverReachable: Boolean,
    val message: String
)

class LocalNetworkPolicy(
    private val context: Context,
    private val serverHost: String,
    private val serverPort: Int
) {

    fun evaluate(): NetworkPolicyResult {

        val connectivityManager =
            context.getSystemService(
                Context.CONNECTIVITY_SERVICE
            ) as ConnectivityManager

        val network =
            connectivityManager.activeNetwork
                ?: return NetworkPolicyResult(
                    allowed = false,
                    wifi = false,
                    localIp = null,
                    subnetOk = false,
                    serverReachable = false,
                    message = "No existe una red activa"
                )

        val capabilities =
            connectivityManager.getNetworkCapabilities(
                network
            )

        val isWifi =
            capabilities?.hasTransport(
                NetworkCapabilities.TRANSPORT_WIFI
            ) == true

        if (!isWifi) {

            return NetworkPolicyResult(
                allowed = false,
                wifi = false,
                localIp = null,
                subnetOk = false,
                serverReachable = false,
                message = "La conexión activa no es Wi-Fi"
            )
        }

        val linkProperties =
            connectivityManager.getLinkProperties(
                network
            )

        val ipv4 =
            linkProperties
                ?.linkAddresses
                ?.map { it.address }
                ?.filterIsInstance<Inet4Address>()
                ?.firstOrNull()

        val ip =
            ipv4?.hostAddress

        val subnetOk =
            ip?.startsWith(
                "192.168.10."
            ) == true

        if (!subnetOk) {

            return NetworkPolicyResult(
                allowed = false,
                wifi = true,
                localIp = ip,
                subnetOk = false,
                serverReachable = false,
                message =
                    "Wi-Fi detectada, pero fuera de 192.168.10.0/24"
            )
        }

        val reachable =
            try {

                Socket().use { socket ->

                    socket.connect(
                        InetSocketAddress(
                            serverHost,
                            serverPort
                        ),
                        2000
                    )
                }

                true

            } catch (_: Exception) {

                false
            }

        return NetworkPolicyResult(
            allowed =
                isWifi &&
                subnetOk &&
                reachable,

            wifi = isWifi,
            localIp = ip,
            subnetOk = subnetOk,
            serverReachable = reachable,

            message =
                if (reachable) {
                    "Red local autorizada"
                } else {
                    "Sync Server no alcanzable"
                }
        )
    }
}
