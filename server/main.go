package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/ROMUALDO-TXT/klever-challenge-golang/database"
	//pb "github.com/ROMUALDO-TXT/klever-challenge-golang/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

var collection *mongo.Collection

//var blogService *pb.BlogServiceServer
var mongoCtx context.Context

var port string = ":" + os.Getenv("SERVER_PORT")

func main() {

	database.CreateConnection()
	mongoCtx = database.GetContext()
	collection = database.GetCollection("posts")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Starting server on port", port)

	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("fail to listen to port %v", port)
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	//pb.

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port ", port)

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	<-c

	fmt.Println("\nStopping the server...")
	server.Stop()
	lis.Close()
	database.CloseConnection()
	fmt.Println("Connection terminated.")

}
