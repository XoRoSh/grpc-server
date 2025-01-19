package main

import (
	"context"
	"log"
	"net"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	pb "github.com/XoRoSh/grpc-server/data"
)

type Data struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Description string
}

type server struct {
	pb.UnimplementedDataServiceServer
	db *gorm.DB
}

func createFieldMask(paths []string) *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{
		Paths: paths,
	}
}

func (s *server) GetData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
	var data Data
	if err := s.db.First(&data, "id = ?", req.Id).Error; err != nil {
		return nil, err
	}

	// Handle FieldMask
	response := &pb.DataResponse{}
	for _, path := range req.FieldMask.GetPaths() {
		// Here dynamically ?
		switch path {
		case "id":
			response.Id = data.ID
		case "name":
			response.Name = data.Name
		case "description":
			response.Description = data.Description
		}
	}

	return response, nil
}

func main() {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&Data{}) // Migrate the schema

	db.Create(&Data{ID: "1", Name: "Example", Description: "This is a sample data entry"})

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDataServiceServer(grpcServer, &server{db: db})

	log.Println("gRPC server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
