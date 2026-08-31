package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/networkpolicy"
)

func main() {

	server :=
		getEnv(
			"SYNC_SERVER",
			"localhost:50055",
		)

	ssid :=
		getEnv(
			"SYNC_CURRENT_SSID",
			"UPB-LAB",
		)

	allowedSSID :=
		getEnv(
			"SYNC_ALLOWED_SSID",
			"UPB-LAB",
		)

	allowedCIDR :=
		getEnv(
			"SYNC_ALLOWED_CIDR",
			"127.0.0.0/8",
		)

	status, err :=
		networkpolicy.Evaluate(
			networkpolicy.Config{
				AllowedSSID: allowedSSID,

				AllowedCIDR: allowedCIDR,

				ServerAddress: server,

				Timeout: 3 * time.Second,
			},
			ssid,
		)

	if err != nil {

		log.Fatalf(
			"Evaluación de red falló: %v",
			err,
		)
	}

	fmt.Println(
		"===================================",
	)

	fmt.Println(
		" NETWORK POLICY - FILE SYNC",
	)

	fmt.Println(
		"===================================",
	)

	fmt.Println(
		"SSID actual:",
		status.SSID,
	)

	fmt.Println(
		"IP local:",
		status.LocalIP,
	)

	fmt.Println(
		"SSID autorizado:",
		status.SSIDAllowed,
	)

	fmt.Println(
		"Subred autorizada:",
		status.NetworkAllowed,
	)

	fmt.Println(
		"Servidor alcanzable:",
		status.ServerReachable,
	)

	fmt.Println(
		"===================================",
	)

	if status.Allowed {

		fmt.Println(
			"DECISIÓN: SYNC PERMITIDO",
		)

		return
	}

	fmt.Println(
		"DECISIÓN: SYNC BLOQUEADO",
	)

	os.Exit(
		2,
	)
}

func getEnv(
	name string,
	fallback string,
) string {

	value :=
		os.Getenv(
			name,
		)

	if value == "" {
		return fallback
	}

	return value
}
