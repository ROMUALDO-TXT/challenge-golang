package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/ROMUALDO-TXT/klever-challenge-golang/database"
	pb "github.com/ROMUALDO-TXT/klever-challenge-golang/proto"
	"github.com/ROMUALDO-TXT/klever-challenge-golang/services"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

var collection *mongo.Collection
var blogService pb.BlogServiceServer
var mongoCtx context.Context

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.CreateConnection()
	mongoCtx = database.GetContext()
	collection = database.GetCollection("posts")
	blogService = services.NewService(collection, mongoCtx)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Starting server on port", os.Getenv("PORT"))

	lis, err := net.Listen("tcp", os.Getenv("PORT"))

	if err != nil {
		log.Fatalf("fail to listen to port %v", os.Getenv("PORT"))
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	pb.RegisterBlogServiceServer(server, blogService)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port ", os.Getenv("PORT"))

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	<-c

	fmt.Println("\nStopping the server...")
	server.Stop()
	lis.Close()
	database.CloseConnection()
	fmt.Println("Connection terminated.")

}
