package workflow

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	workflowv1 "github.com/gobenpark/colign/gen/proto/workflow/v1"
	"github.com/gobenpark/colign/gen/proto/workflow/v1/workflowv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobenpark/colign/internal/models"
	"github.com/uptrace/bun"
)

type ConnectHandler struct {
	service *Service
	db      *bun.DB
}

var _ workflowv1connect.WorkflowServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, db *bun.DB) *ConnectHandler {
	return &ConnectHandler{service: service, db: db}
}

func (h *ConnectHandler) GetStatus(ctx context.Context, req *connect.Request[workflowv1.GetStatusRequest]) (*connect.Response[workflowv1.GetStatusResponse], error) {
	stage, conditions, err := h.service.GetStatus(ctx, req.Msg.ChangeId)
	if err != nil {
		if errors.Is(err, ErrChangeNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoConditions := make([]*workflowv1.GateCondition, len(conditions))
	for i, c := range conditions {
		protoConditions[i] = &workflowv1.GateCondition{
			Name:        c.Name,
			Description: c.Description,
			Met:         c.Met,
		}
	}

	return connect.NewResponse(&workflowv1.GetStatusResponse{
		Stage:      string(stage),
		Conditions: protoConditions,
	}), nil
}

func (h *ConnectHandler) Approve(ctx context.Context, req *connect.Request[workflowv1.ApproveRequest]) (*connect.Response[workflowv1.ApproveResponse], error) {
	userID := int64(1) // TODO: from auth context

	approval := &models.Approval{
		ChangeID: req.Msg.ChangeId,
		UserID:   userID,
		Status:   "approved",
		Comment:  req.Msg.Comment,
	}
	if _, err := h.db.NewInsert().Model(approval).
		On("CONFLICT (change_id, user_id) DO UPDATE").
		Set("status = EXCLUDED.status").
		Set("comment = EXCLUDED.comment").
		Exec(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	advanced, err := h.service.EvaluateAndAdvance(ctx, req.Msg.ChangeId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &workflowv1.ApproveResponse{Advanced: advanced}
	if advanced {
		change := new(models.Change)
		if err := h.db.NewSelect().Model(change).Where("id = ?", req.Msg.ChangeId).Scan(ctx); err == nil {
			resp.NewStage = string(change.Stage)
		}
	}

	return connect.NewResponse(resp), nil
}

func (h *ConnectHandler) RequestChanges(ctx context.Context, req *connect.Request[workflowv1.RequestChangesRequest]) (*connect.Response[workflowv1.RequestChangesResponse], error) {
	userID := int64(1) // TODO: from auth context

	if err := h.service.Revert(ctx, req.Msg.ChangeId, userID, req.Msg.Reason); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	change := new(models.Change)
	h.db.NewSelect().Model(change).Where("id = ?", req.Msg.ChangeId).Scan(ctx)

	return connect.NewResponse(&workflowv1.RequestChangesResponse{
		NewStage: string(change.Stage),
	}), nil
}

func (h *ConnectHandler) Revert(ctx context.Context, req *connect.Request[workflowv1.RevertRequest]) (*connect.Response[workflowv1.RevertResponse], error) {
	userID := int64(1) // TODO: from auth context

	if err := h.service.Revert(ctx, req.Msg.ChangeId, userID, req.Msg.Reason); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	change := new(models.Change)
	h.db.NewSelect().Model(change).Where("id = ?", req.Msg.ChangeId).Scan(ctx)

	return connect.NewResponse(&workflowv1.RevertResponse{
		NewStage: string(change.Stage),
	}), nil
}

func (h *ConnectHandler) GetHistory(ctx context.Context, req *connect.Request[workflowv1.GetHistoryRequest]) (*connect.Response[workflowv1.GetHistoryResponse], error) {
	var events []models.WorkflowEvent
	err := h.db.NewSelect().Model(&events).
		Where("change_id = ?", req.Msg.ChangeId).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoEvents := make([]*workflowv1.WorkflowEvent, len(events))
	for i, e := range events {
		protoEvents[i] = &workflowv1.WorkflowEvent{
			Id:        e.ID,
			FromStage: e.FromStage,
			ToStage:   e.ToStage,
			Action:    e.Action,
			Reason:    e.Reason,
			UserId:    e.UserID,
			CreatedAt: timestamppb.New(e.CreatedAt),
		}
	}

	return connect.NewResponse(&workflowv1.GetHistoryResponse{
		Events: protoEvents,
	}), nil
}

func (h *ConnectHandler) SetApprovalPolicy(ctx context.Context, req *connect.Request[workflowv1.SetApprovalPolicyRequest]) (*connect.Response[workflowv1.SetApprovalPolicyResponse], error) {
	policy := &models.ApprovalPolicy{
		ProjectID: req.Msg.ProjectId,
		Policy:    req.Msg.Policy,
		MinCount:  int(req.Msg.MinCount),
	}

	if _, err := h.db.NewInsert().Model(policy).
		On("CONFLICT (project_id) DO UPDATE").
		Set("policy = EXCLUDED.policy").
		Set("min_count = EXCLUDED.min_count").
		Set("updated_at = NOW()").
		Exec(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflowv1.SetApprovalPolicyResponse{}), nil
}
