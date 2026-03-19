package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	authv1 "github.com/gobenpark/colign/gen/proto/auth/v1"
	"github.com/gobenpark/colign/gen/proto/auth/v1/authv1connect"
)

type ConnectHandler struct {
	service      *Service
	oauthService *OAuthService
}

var _ authv1connect.AuthServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, oauthService *OAuthService) *ConnectHandler {
	return &ConnectHandler{service: service, oauthService: oauthService}
}

func (h *ConnectHandler) Register(ctx context.Context, req *connect.Request[authv1.RegisterRequest]) (*connect.Response[authv1.RegisterResponse], error) {
	tokenPair, err := h.service.Register(ctx, RegisterRequest{
		Email:    req.Msg.Email,
		Password: req.Msg.Password,
		Name:     req.Msg.Name,
	})
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.RegisterResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}), nil
}

func (h *ConnectHandler) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	tokenPair, err := h.service.Login(ctx, LoginRequest{
		Email:    req.Msg.Email,
		Password: req.Msg.Password,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}), nil
}

func (h *ConnectHandler) RefreshToken(ctx context.Context, req *connect.Request[authv1.RefreshTokenRequest]) (*connect.Response[authv1.RefreshTokenResponse], error) {
	tokenPair, err := h.service.RefreshToken(ctx, req.Msg.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}), nil
}

func (h *ConnectHandler) VerifyEmail(ctx context.Context, req *connect.Request[authv1.VerifyEmailRequest]) (*connect.Response[authv1.VerifyEmailResponse], error) {
	if err := h.service.VerifyEmail(ctx, req.Msg.Token); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&authv1.VerifyEmailResponse{
		Message: "email verified",
	}), nil
}

func (h *ConnectHandler) GetOAuthURL(ctx context.Context, req *connect.Request[authv1.GetOAuthURLRequest]) (*connect.Response[authv1.GetOAuthURLResponse], error) {
	url, err := h.oauthService.GetAuthURL(req.Msg.Provider, "state") // TODO: proper state management
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&authv1.GetOAuthURLResponse{
		Url: url,
	}), nil
}

func (h *ConnectHandler) OAuthCallback(ctx context.Context, req *connect.Request[authv1.OAuthCallbackRequest]) (*connect.Response[authv1.OAuthCallbackResponse], error) {
	tokenPair, err := h.oauthService.HandleCallback(ctx, req.Msg.Provider, req.Msg.Code)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.OAuthCallbackResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}), nil
}
