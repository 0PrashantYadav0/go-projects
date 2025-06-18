package main

import (
	"context"
	pb "grpc-go-project/proto"
	"log"
	"time"
)

func callSayHello(client pb.GreetServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.SayHello(ctx, &pb.NoParams{})
	if err != nil {
		log.Fatalf("could not greet : %v", err)
	}
	log.Printf("%s", res.Message)
}
