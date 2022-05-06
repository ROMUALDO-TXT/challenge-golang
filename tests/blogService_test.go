package tests

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/ROMUALDO-TXT/klever-challenge-golang/database"
	pb "github.com/ROMUALDO-TXT/klever-challenge-golang/proto"
	"github.com/ROMUALDO-TXT/klever-challenge-golang/services"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type blogTest struct {
	Id        string
	AuthorId  string
	Content   string
	Title     string
	Upvotes   int64
	Downvotes int64
	Score     int64
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var collection *mongo.Collection
var blogService pb.BlogServiceServer
var mongoCtx context.Context

func init() {
	os.Setenv("MONGO_DB_URL", "mongodb+srv://goblogtestuser:kKJ7oj3f0S3IuFJD@cluster0.cyk1r.mongodb.net/Klever-blog?retryWrites=true&w=majority")

	database.CreateConnection()

	mongoCtx = database.GetContext()
	collection = database.GetCollection("posts-test")
	blogService = services.NewService(collection, mongoCtx)

	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterBlogServiceServer(server, blogService)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func CreateBlogTest(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()

	client := pb.NewBlogServiceClient(conn)

	req := pb.CreateBlogReq{
		Title:    "test-title",
		Content:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		AuthorId: "test-author1",
	}

	r, err := client.CreateBlog(ctx, &req)
	if err != nil {
		t.Fatalf("Create Blog failed: %v", err)
	}

	assert.Equal(t, r.GetBlog().GetTitle(), req.Title)
	assert.Equal(t, r.GetBlog().GetContent(), req.Content)
	assert.Equal(t, r.GetBlog().GetAuthorId(), req.AuthorId)
}

func TestUpvote(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	client := pb.NewBlogServiceClient(conn)

	req := pb.UpvoteReq{Id: "6272dde41efd65dfdfc5c9d9"}
	r, err := client.Upvote(ctx, &req)
	if err != nil {
		t.Fatalf("UpvoteBlog failed: %v", err)
	}

	assert.Equal(t, r.Success, true)
}

func TestDownvote(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewBlogServiceClient(conn)

	req := pb.DownvoteReq{Id: "6272dde41efd65dfdfc5c9d9"}
	r, err := client.Downvote(ctx, &req)
	if err != nil {
		t.Fatalf("Downvote failed: %v", err)
	}

	assert.Equal(t, r.Success, true)
}

func TestDeleteBlog(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewBlogServiceClient(conn)

	createReq := pb.CreateBlogReq{
		Title:    "test-delete-title",
		Content:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		AuthorId: "test-author1",
	}

	createRes, err := client.CreateBlog(ctx, &createReq)

	req := pb.DeleteBlogReq{Id: createRes.Blog.Id}
	r, err := client.DeleteBlog(ctx, &req)
	if err != nil {
		t.Fatalf("Delete Blog failed: %v", err)
	}

	assert.Equal(t, r.Success, true)
}

func ReadBlogTest(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewBlogServiceClient(conn)

	mock := blogTest{
		Upvotes:   0,
		Downvotes: 0,
		Score:     0,
		Title:     "test-read",
		Content:   "lorem ipsum dolor",
		AuthorId:  "test-author2",
		Id:        "627414477592409bf2943045",
	}

	req := pb.ReadBlogReq{Id: "627414477592409bf2943045"}
	resp, err := client.ReadBlog(ctx, &req)
	if err != nil {
		t.Fatalf("ReadBlog failed: %v", err)
	}

	assert.Equal(t, resp.Blog.Id, mock.Id)
	assert.Equal(t, resp.Blog.Title, mock.Title)
	assert.Equal(t, resp.Blog.Content, mock.Content)
	assert.Equal(t, resp.Blog.AuthorId, mock.AuthorId)
	assert.Equal(t, resp.Blog.Upvotes, mock.Upvotes)
	assert.Equal(t, resp.Blog.Downvotes, mock.Downvotes)
	assert.Equal(t, resp.Blog.Score, mock.Score)

}
