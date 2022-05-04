package services

import (
	"context"
	"fmt"

	models "github.com/ROMUALDO-TXT/klever-challenge-golang/models"
	pb "github.com/ROMUALDO-TXT/klever-challenge-golang/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type blogServiceServer struct {
	collection *mongo.Collection
	mongoCtx   context.Context
	pb.UnimplementedBlogServiceServer
}

func NewService(collection *mongo.Collection, mongoCtx context.Context) pb.BlogServiceServer {
	return &blogServiceServer{
		collection: collection,
		mongoCtx:   mongoCtx,
	}
}

func (server *blogServiceServer) CreateBlog(ctx context.Context, req *pb.CreateBlogReq) (*pb.BlogRes, error) {
	//Validation
	if req.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Verify the fields!"))
	}
	if req.GetContent() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Verify the fields!"))
	}
	if req.GetAuthorId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Author ID can1t be empty!"))
	}

	data := models.Blog{
		AuthorId:  req.GetAuthorId(),
		Title:     req.GetTitle(),
		Content:   req.GetContent(),
		Upvotes:   0,
		Downvotes: 0,
		Score:     0,
	}

	result, err := server.collection.InsertOne(server.mongoCtx, data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid := result.InsertedID.(primitive.ObjectID)

	response := &pb.BlogRes{
		Blog: &pb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorId,
			Title:    data.Title,
			Content:  data.Content,
			Upvote:   data.Upvotes,
			Downvote: data.Downvotes,
			Score:    data.Score,
		},
	}

	return response, nil
}

func (server *blogServiceServer) DeleteBlog(ctx context.Context, req *pb.DeleteBlogReq) (*pb.SuccessRes, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Verify the fields!"))
	}

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	_, err = server.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	return &pb.SuccessRes{
		Success: true,
	}, nil
}

func (server *blogServiceServer) ListBlogs(req *pb.ListBlogsReq, stream pb.BlogService_ListBlogsServer) error {

	data := &models.Blog{}

	cursor, err := server.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		if err := cursor.Decode(data); err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}

		stream.Send(&pb.BlogRes{
			Blog: &pb.Blog{
				Id:       data.Id.Hex(),
				AuthorId: data.AuthorId,
				Content:  data.Content,
				Title:    data.Title,
			},
		})
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unkown cursor error: %v", err))
	}
	return nil
}

func (server *blogServiceServer) Upvote(ctx context.Context, req *pb.UpvoteReq) (*pb.SuccessRes, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Verify the fields!"))
	}

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied blog id to a MongoDB ObjectId: %v", err),
		)
	}

	filter := bson.M{"_id": oid}

	_, err = server.collection.UpdateOne(ctx, filter, bson.M{"$inc": bson.M{"upvotes": 1, "score": 1}}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find a blog with id %s: %v", req.GetId(), err))
	}

	return &pb.SuccessRes{
		Success: true,
	}, nil
}

func (server *blogServiceServer) Downvote(ctx context.Context, req *pb.DownvoteReq) (*pb.SuccessRes, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Verify the fields!"))
	}

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert the supplied blog id to a MongoDB ObjectId: %v", err),
		)
	}

	filter := bson.M{"_id": oid}

	_, err = server.collection.UpdateOne(ctx, filter, bson.M{"$inc": bson.M{"downvotes": 1, "score": -1}}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find a blog with id %s: %v", req.GetId(), err))
	}

	return &pb.SuccessRes{
		Success: true,
	}, nil
}
