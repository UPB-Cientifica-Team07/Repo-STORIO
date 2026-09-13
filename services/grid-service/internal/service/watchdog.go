package service

import (
	"context"
	"log"
	"time"
)

func RunWatchdog(
	ctx context.Context,
	service *GridService,
	interval time.Duration,
	timeout time.Duration,
) {

	ticker :=
		time.NewTicker(
			interval,
		)

	defer ticker.Stop()

	run :=
		func() {

			affected, err :=
				service.MarkInactive(
					timeout,
				)

			if err != nil {

				log.Printf(
					"[GRID WATCHDOG] error: %v",
					err,
				)

				return
			}

			if affected > 0 {

				log.Printf(
					"[GRID WATCHDOG] recursos marcados INACTIVE=%d",
					affected,
				)
			}
		}

	run()

	for {

		select {

		case <-ctx.Done():

			log.Println(
				"[GRID WATCHDOG] detenido",
			)

			return

		case <-ticker.C:

			run()
		}
	}
}
