package main

import (
	"context"
	"fmt"
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	pb "github.com/XoRoSh/grpc-server/data"
)

type server struct {
	pb.UnimplementedDataServiceServer
	db *gorm.DB
}

func (s *server) GetData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
	structureMap := map[string]interface{}{
		"Data": Data{},
		"Car":  Car{},
	}

	structure, ok := structureMap[req.Id]
	if !ok {
		return nil, fmt.Errorf("unknown structure: %s", req.Id)
	}

	dataPtr := reflect.New(reflect.TypeOf(structure)).Interface()

	if err := s.db.First(dataPtr, "id = ?", req.Id).Error; err != nil {
		return nil, err
	}

	response := &pb.DataResponse{}
	val := reflect.ValueOf(dataPtr).Elem()

	// Populate fields based on the field mask
	for _, path := range req.FieldMask.GetPaths() {
		field := val.FieldByName(cases.Title(language.Und).String(path))
		if field.IsValid() {
			switch field.Kind() {
			case reflect.String:
				reflect.ValueOf(response).Elem().FieldByName(cases.Title(language.Und).String(path)).SetString(field.String())
			}
		}
	}

	return response, nil
}

func NewServer(db *gorm.DB) *server {
	return &server{db: db}
}

func RegisterServices(s *grpc.Server, srv *server) {
	pb.RegisterDataServiceServer(s, srv)
}
