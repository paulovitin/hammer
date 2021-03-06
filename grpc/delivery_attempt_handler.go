package grpc

import (
	"context"
	"database/sql"

	"github.com/allisson/hammer"
	pb "github.com/allisson/hammer/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeliveryAttemptHandler implements methods for DeliveryAttempt get/list
type DeliveryAttemptHandler struct {
	deliveryAttemptService hammer.DeliveryAttemptService
}

func (d *DeliveryAttemptHandler) buildResponse(deliveryAttempt *hammer.DeliveryAttempt) (*pb.DeliveryAttempt, error) {
	response := &pb.DeliveryAttempt{}
	createdAt, err := ptypes.TimestampProto(deliveryAttempt.CreatedAt)
	if err != nil {
		return response, status.Error(codes.Internal, err.Error())
	}
	response.Id = deliveryAttempt.ID
	response.DeliveryId = deliveryAttempt.DeliveryID
	response.Request = deliveryAttempt.Request
	response.Response = deliveryAttempt.Response
	response.ResponseStatusCode = uint32(deliveryAttempt.ResponseStatusCode)
	response.ExecutionDuration = uint32(deliveryAttempt.ExecutionDuration)
	response.Success = deliveryAttempt.Success
	response.Error = deliveryAttempt.Error
	response.CreatedAt = createdAt

	return response, nil
}

// GetDeliveryAttempt gets the DeliveryAttempt
func (d *DeliveryAttemptHandler) GetDeliveryAttempt(ctx context.Context, request *pb.GetDeliveryAttemptRequest) (*pb.DeliveryAttempt, error) {
	// Get DeliveryAttempt from service
	deliveryAttempt, err := d.deliveryAttemptService.Find(request.Id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &pb.DeliveryAttempt{}, status.Error(codes.NotFound, hammer.ErrDeliveryAttemptDoesNotExists.Error())
		default:
			return &pb.DeliveryAttempt{}, status.Error(codes.Internal, err.Error())
		}
	}

	return d.buildResponse(&deliveryAttempt)
}

// ListDeliveryAttempts get a list of DeliveryAttempts
func (d *DeliveryAttemptHandler) ListDeliveryAttempts(ctx context.Context, request *pb.ListDeliveryAttemptsRequest) (*pb.ListDeliveryAttemptsResponse, error) {
	// Get limit and offset
	limit, offset := parsePagination(request.Limit, request.Offset)

	// Create response
	response := &pb.ListDeliveryAttemptsResponse{}

	// Get DeliveryAttempts
	findOptions := hammer.FindOptions{
		FindPagination: &hammer.FindPagination{
			Limit:  limit,
			Offset: offset,
		},
	}
	if request.DeliveryId != "" {
		deliveryFilter := hammer.FindFilter{
			FieldName: "delivery_id",
			Operator:  "=",
			Value:     request.DeliveryId,
		}
		findOptions.FindFilters = append(findOptions.FindFilters, deliveryFilter)
	}
	createdAtFilters := createdAtFilters(request.CreatedAtGt, request.CreatedAtGte, request.CreatedAtLt, request.CreatedAtLte)
	findOptions.FindFilters = append(findOptions.FindFilters, createdAtFilters...)
	deliveryAttempts, err := d.deliveryAttemptService.FindAll(findOptions)
	if err != nil {
		return response, status.Error(codes.Internal, err.Error())
	}

	// Update response
	for _, deliveryAttempt := range deliveryAttempts {
		deliveryAttemptResponse, err := d.buildResponse(&deliveryAttempt)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
		response.DeliveryAttempts = append(response.DeliveryAttempts, deliveryAttemptResponse)
	}

	return response, nil
}

// NewDeliveryAttemptHandler returns a new DeliveryAttempt
func NewDeliveryAttemptHandler(deliveryAttemptService hammer.DeliveryAttemptService) DeliveryAttemptHandler {
	return DeliveryAttemptHandler{deliveryAttemptService: deliveryAttemptService}
}
