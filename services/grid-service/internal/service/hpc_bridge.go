package service

import (
	"context"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/repository"
)

func RunHpcResourceBridge(
	ctx context.Context,
	resourceRepository *repository.ResourceRepository,
	interval time.Duration,
	timeout time.Duration,
) {

	if interval <= 0 {
		log.Println(
			"[GRID HPC BRIDGE] intervalo inválido",
		)

		return
	}

	if timeout <= 0 {
		log.Println(
			"[GRID HPC BRIDGE] timeout inválido",
		)

		return
	}

	syncResources :=
		func() {

			cutoff :=
				time.Now().
					Add(
						-timeout,
					)

			affected, err :=
				resourceRepository.
					SyncHpcResources(
						cutoff,
					)

			if err != nil {

				log.Printf(
					"[GRID HPC BRIDGE] error: %v",
					err,
				)

				return
			}

			log.Printf(
				"[GRID HPC BRIDGE] recursos HPC sincronizados=%d",
				affected,
			)
		}

	// Sincronización inicial.
	syncResources()

	ticker :=
		time.NewTicker(
			interval,
		)

	defer ticker.Stop()

	for {

		select {

		case <-ctx.Done():

			log.Println(
				"[GRID HPC BRIDGE] detenido",
			)

			return

		case <-ticker.C:

			syncResources()
		}
	}
}
