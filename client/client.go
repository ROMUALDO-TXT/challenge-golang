package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ROMUALDO-TXT/klever-challenge-golang/models"
	pb "github.com/ROMUALDO-TXT/klever-challenge-golang/proto"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var client pb.BlogServiceClient

var blogs []models.Blog

func main() {

	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := grpc.Dial(os.Getenv("SERVER_URL"), grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	client = pb.NewBlogServiceClient(conn)

	g := gin.Default()
	g.POST("/blog", createBlog)
	g.POST("/upvote/:id", upvoteBlog)
	g.POST("/downvote/:id", downvoteBlog)
	g.GET("/blogs", listBlogs)
	g.GET("/blog/id", readBlog)
	g.DELETE("/blog/delete/:id", deleteBlog)

	log.Fatal(g.Run(":8080"))
}

func createBlog(ctx *gin.Context) {

	blog := pb.CreateBlogReq{}
	err := ctx.ShouldBindJSON(&blog)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := client.CreateBlog(ctx, &blog)
	if err == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"result": fmt.Sprint(res),
		})
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	fmt.Println(res)
}
func upvoteBlog(ctx *gin.Context) {
	uid := ctx.Param("id")

	obj := pb.UpvoteReq{Id: uid}

	res, err := client.Upvote(ctx, &obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func downvoteBlog(ctx *gin.Context) {
	uid := ctx.Param("id")

	obj := pb.DownvoteReq{Id: uid}

	res, err := client.Downvote(ctx, &obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func readBlog(ctx *gin.Context) {
	id := ctx.Param("id")

	obj := pb.ReadBlogReq{Id: id}

	res, err := client.ReadBlog(ctx, &obj)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func listBlogs(ctx *gin.Context) {

	obj := pb.ListBlogsReq{}

	stream, err := client.ListBlogs(ctx, &obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for {

		res, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println(res.GetBlog())
		ctx.JSON(http.StatusOK, res.GetBlog())
	}

}

func deleteBlog(ctx *gin.Context) {

	uid := ctx.Param("id")

	obj := pb.DeleteBlogReq{Id: uid}

	res, err := client.DeleteBlog(ctx, &obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
