package memory

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	memoryv1 "github.com/gobenpark/colign/gen/proto/memory/v1"
	"github.com/gobenpark/colign/gen/proto/memory/v1/memoryv1connect"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

type ConnectHandler struct {
	service    *Service
	jwtManager *auth.JWTManager
}

var _ memoryv1connect.MemoryServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, jwtManager *auth.JWTManager) *ConnectHandler {
	return &ConnectHandler{service: service, jwtManager: jwtManager}
}

func (h *ConnectHandler) extractClaims(header string) (*auth.Claims, error) {
	claims, err := auth.ExtractClaims(h.jwtManager, header)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	return claims, nil
}

func (h *ConnectHandler) GetMemory(ctx context.Context, req *connect.Request[memoryv1.GetMemoryRequest]) (*connect.Response[memoryv1.GetMemoryResponse], error) {
	_, err := h.extractClaims(req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	mem, err := h.service.Get(ctx, req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &memoryv1.GetMemoryResponse{}
	if mem != nil {
		resp.Memory = memoryToProto(mem)
	}

	return connect.NewResponse(resp), nil
}

func (h *ConnectHandler) SaveMemory(ctx context.Context, req *connect.Request[memoryv1.SaveMemoryRequest]) (*connect.Response[memoryv1.SaveMemoryResponse], error) {
	claims, err := h.extractClaims(req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	mem, err := h.service.Save(ctx, req.Msg.ProjectId, req.Msg.Content, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&memoryv1.SaveMemoryResponse{
		Memory: memoryToProto(mem),
	}), nil
}

func memoryToProto(m *models.ProjectMemory) *memoryv1.ProjectMemory {
	proto := &memoryv1.ProjectMemory{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		Content:   m.Content,
		UpdatedBy: m.UpdatedBy,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
	if m.User != nil {
		proto.UpdatedByName = m.User.Name
	}
	return proto
}
