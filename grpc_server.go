// package main

// import (
// 	"context"
// 	"log"
// 	"net"

// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"

// 	"google.golang.org/grpc"
// 	"google.golang.org/protobuf/types/known/fieldmaskpb"

// 	pb "github.com/XoRoSh/grpc-server/data"
// )

// type Data struct {
// 	ID          string `gorm:"primaryKey"`
// 	Name        string
// 	Description string
// 	Car         Car
// }

// type Car struct {
// 	VIN   string `gorm:"primaryKey"`
// 	Color string
// 	Fuel  string
// }

// type server struct {
// 	pb.UnimplementedDataServiceServer
// 	db *gorm.DB
// }

// func createFieldMask(paths []string) *fieldmaskpb.FieldMask {
// 	return &fieldmaskpb.FieldMask{
// 		Paths: paths,
// 	}
// }

// func (s *server) GetData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
// 	var data Data
// 	if err := s.db.First(&data, "id = ?", req.Id).Error; err != nil {
// 		return nil, err
// 	}

// 	// Handle FieldMask
// 	response := &pb.DataResponse{}
// 	for _, path := range req.FieldMask.GetPaths() {
// 		switch path {
// 		case "id":
// 			response.Id = data.ID
// 		case "name":
// 			response.Name = data.Name
// 		case "description":
// 			response.Description = data.Description
// 		}
// 	}
// 	return response, nil
// }

// func main() {
// 	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	db.AutoMigrate(&Data{}) // Migrate the schema

// 	// Sample data
// 	db.Create(&Data{ID: "1", Name: "Example", Description: "This is a sample data entry"})
// 	db.Create(&Data{ID: "2", Name: "Example2 ", Description: "This is a sample data entry 2"})
// 	db.Create(&Data{ID: "3", Name: "Example3 ", Description: "This is a sample data entry 3"})
// 	db.Create(&Data{ID: "4", Name: "Example4 ", Description: "This is a sample data entry 4"})
// 	db.Create(&Data{ID: "5", Name: "Example4 ", Description: "This is a sample data entry 4"})
// 	db.Create(&Data{ID: "6", Name: "Example4 ", Description: "This is a sample data entry 4"})

// 	// Start gRPC server
// 	listener, err := net.Listen("tcp", ":50051")
// 	if err != nil {
// 		log.Fatalf("Failed to listen: %v", err)
// 	}
// 	grpcServer := grpc.NewServer()
// 	pb.RegisterDataServiceServer(grpcServer, &server{db: db})

// 	log.Println("gRPC server is running on port 50051")
// 	if err := grpcServer.Serve(listener); err != nil {
// 		log.Fatalf("Failed to serve: %v", err)
// 	}
// }

package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	pb "github.com/XoRoSh/grpc-server/data"
	"github.com/graphql-go/graphql"
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

// GraphQL Schema Definition
func createGraphQLSchema(db *gorm.DB) graphql.Schema {
	// Define GraphQL types
	dataType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Data",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.String},
			"name":        &graphql.Field{Type: graphql.String},
			"description": &graphql.Field{Type: graphql.String},
		},
	})

	// Define the root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"getItems": &graphql.Field{
				Type: graphql.NewList(dataType), // This should return a list of items
				Args: graphql.FieldConfigArgument{
					"id":          &graphql.ArgumentConfig{Type: graphql.String},
					"name":        &graphql.ArgumentConfig{Type: graphql.String},
					"description": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					filters := map[string]interface{}{}
					if id, ok := params.Args["id"].(string); ok {
						filters["id"] = id
					}
					if name, ok := params.Args["name"].(string); ok {
						filters["name"] = name
					}
					if description, ok := params.Args["description"].(string); ok {
						filters["description"] = description
					}

					var items []Data
					if err := db.Where(filters).Find(&items).Error; err != nil {
						return nil, err
					}
					return items, nil
				},
			},
		},
	})

	// Create the schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}
	return schema
}

func main() {
	// Initialize the database
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&Data{}) // Migrate the schema

	// Seed sample data
	db.Create(&Data{ID: "1", Name: "Example", Description: "This is a sample data entry"})
	db.Create(&Data{ID: "2", Name: "Example2", Description: "This is a sample data entry 2"})
	db.Create(&Data{ID: "3", Name: "Example3", Description: "This is a sample data entry 3"})
	db.Create(&Data{ID: "4", Name: "Example4", Description: "This is a sample data entry 4"})
	db.Create(&Data{ID: "5", Name: "Example4", Description: "This is a sample data entry 5"})
	db.Create(&Data{ID: "6", Name: "Example4", Description: "This is a sample data entry 4"})

	// Start gRPC server
	go func() {
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
	}()

	// Start GraphQL server
	schema := createGraphQLSchema(db)
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
		})
		if len(result.Errors) > 0 {
			log.Printf("Failed to execute query: %v", result.Errors)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})

	log.Println("GraphQL server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
