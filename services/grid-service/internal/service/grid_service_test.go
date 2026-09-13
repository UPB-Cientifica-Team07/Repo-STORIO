package service

import (
	"testing"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/model"
)

func TestValidateDeviceType(
	t *testing.T,
) {

	valid :=
		[]string{
			model.DeviceTypeCompute,
			model.DeviceTypeMobile,
			model.DeviceTypeIoT,
		}

	for _, value := range valid {

		if err :=
			validateDeviceType(
				value,
			); err != nil {

			t.Fatalf(
				"tipo válido rechazado %s: %v",
				value,
				err,
			)
		}
	}

	if err :=
		validateDeviceType(
			"INVALID",
		); err == nil {

		t.Fatal(
			"tipo inválido fue aceptado",
		)
	}
}

func TestValidateStatus(
	t *testing.T,
) {

	valid :=
		[]string{
			model.StatusAvailable,
			model.StatusBusy,
			model.StatusInactive,
		}

	for _, value := range valid {

		if err :=
			validateStatus(
				value,
			); err != nil {

			t.Fatalf(
				"estado válido rechazado %s: %v",
				value,
				err,
			)
		}
	}

	if err :=
		validateStatus(
			"INVALID",
		); err == nil {

		t.Fatal(
			"estado inválido fue aceptado",
		)
	}
}

func TestValidateHeartbeatStatus(
	t *testing.T,
) {

	if err :=
		validateHeartbeatStatus(
			model.StatusAvailable,
		); err != nil {

		t.Fatal(err)
	}

	if err :=
		validateHeartbeatStatus(
			model.StatusBusy,
		); err != nil {

		t.Fatal(err)
	}

	if err :=
		validateHeartbeatStatus(
			model.StatusInactive,
		); err == nil {

		t.Fatal(
			"heartbeat no debe aceptar INACTIVE",
		)
	}
}
