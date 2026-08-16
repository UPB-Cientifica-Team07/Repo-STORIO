package main

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const authAddress = "localhost:50052"

func main() {

	conn, err := grpc.NewClient(
		authAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar con Auth Service: %v",
			err,
		)
	}

	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	// ===================================
	// REGISTRO
	// ===================================

	log.Println("===================================")
	log.Println(" REGISTRANDO USUARIO")
	log.Println("===================================")

	registerResponse, err := client.Register(
		ctx,
		&pb.RegisterRequest{
			Username: "Samuel",
			Email:    "samuel@example.com",
			Password: "123456",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error registrando usuario: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		registerResponse.GetSuccess(),
	)

	log.Printf(
		"Mensaje: %s",
		registerResponse.GetMessage(),
	)

	log.Printf(
		"User ID: %s",
		registerResponse.GetUserId(),
	)

	// ===================================
	// LOGIN
	// ===================================

	log.Println("===================================")
	log.Println(" INICIANDO SESIÓN")
	log.Println("===================================")

	loginResponse, err := client.Login(
		ctx,
		&pb.LoginRequest{
			Email:    "samuel@example.com",
			Password: "123456",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error iniciando sesión: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		loginResponse.GetSuccess(),
	)

	log.Printf(
		"Mensaje: %s",
		loginResponse.GetMessage(),
	)

	log.Printf(
		"User ID: %s",
		loginResponse.GetUserId(),
	)

	log.Printf(
		"Token: %s",
		loginResponse.GetToken(),
	)

	// ===================================
	// VALIDAR TOKEN
	// ===================================

	log.Println("===================================")
	log.Println(" VALIDANDO TOKEN")
	log.Println("===================================")

	tokenResponse, err := client.ValidateToken(
		ctx,
		&pb.TokenRequest{
			Token: loginResponse.GetToken(),
		},
	)

	if err != nil {
		log.Fatalf(
			"Error validando token: %v",
			err,
		)
	}

	log.Printf(
		"Válido: %t",
		tokenResponse.GetValid(),
	)

	log.Printf(
		"Mensaje: %s",
		tokenResponse.GetMessage(),
	)

	log.Printf(
		"User ID: %s",
		tokenResponse.GetUserId(),
	)
}
