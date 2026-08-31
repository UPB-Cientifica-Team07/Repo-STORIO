package networkpolicy

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Config struct {
	AllowedSSID string

	AllowedCIDR string

	ServerAddress string

	Timeout time.Duration
}

type Status struct {
	SSID string

	LocalIP string

	SSIDAllowed bool

	NetworkAllowed bool

	ServerReachable bool

	Allowed bool
}

func Evaluate(
	config Config,
	currentSSID string,
) (
	Status,
	error,
) {

	status :=
		Status{
			SSID: strings.TrimSpace(
				currentSSID,
			),
		}

	if config.Timeout <= 0 {
		config.Timeout =
			3 * time.Second
	}

	// =====================================
	// SSID
	// =====================================

	if strings.TrimSpace(
		config.AllowedSSID,
	) == "" {

		status.SSIDAllowed =
			true

	} else {

		status.SSIDAllowed =
			strings.EqualFold(
				strings.TrimSpace(
					config.AllowedSSID,
				),
				status.SSID,
			)
	}

	// =====================================
	// RED LOCAL
	// =====================================

	localIP, err :=
		discoverLocalIP(
			config.ServerAddress,
		)

	if err != nil {

		return status,
			fmt.Errorf(
				"no se pudo determinar IP local: %w",
				err,
			)
	}

	status.LocalIP =
		localIP.String()

	_, network, err :=
		net.ParseCIDR(
			strings.TrimSpace(
				config.AllowedCIDR,
			),
		)

	if err != nil {

		return status,
			fmt.Errorf(
				"CIDR inválido %q: %w",
				config.AllowedCIDR,
				err,
			)
	}

	status.NetworkAllowed =
		network.Contains(
			localIP,
		)

	// =====================================
	// SERVIDOR
	// =====================================

	connection, err :=
		net.DialTimeout(
			"tcp",
			config.ServerAddress,
			config.Timeout,
		)

	if err == nil {

		status.ServerReachable =
			true

		_ =
			connection.Close()
	}

	// =====================================
	// DECISIÓN FINAL
	// =====================================

	status.Allowed =
		status.SSIDAllowed &&
			status.NetworkAllowed &&
			status.ServerReachable

	return status, nil
}

func discoverLocalIP(
	serverAddress string,
) (
	net.IP,
	error,
) {

	host, _, err :=
		net.SplitHostPort(
			serverAddress,
		)

	if err != nil {
		return nil, err
	}

	serverIPs, err :=
		net.LookupIP(
			host,
		)

	if err != nil {
		return nil, err
	}

	var targetIP net.IP

	for _, candidate := range serverIPs {

		if candidate.To4() != nil {

			targetIP =
				candidate

			break
		}
	}

	if targetIP == nil {

		return nil,
			fmt.Errorf(
				"servidor sin dirección IPv4",
			)
	}

	connection, err :=
		net.DialUDP(
			"udp",
			nil,
			&net.UDPAddr{
				IP: targetIP,

				Port: 9,
			},
		)

	if err != nil {
		return nil, err
	}

	defer connection.Close()

	localAddress, ok :=
		connection.LocalAddr().(*net.UDPAddr)

	if !ok {

		return nil,
			fmt.Errorf(
				"dirección local no reconocida",
			)
	}

	if localAddress.IP == nil {

		return nil,
			fmt.Errorf(
				"IP local no disponible",
			)
	}

	return localAddress.IP,
		nil
}
