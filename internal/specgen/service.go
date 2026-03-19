package specgen

import (
	"context"
	"strings"

	"github.com/gobenpark/colign/internal/chat"
)

type Service struct {
	claude *chat.ClaudeClient
}

func NewService(claude *chat.ClaudeClient) *Service {
	return &Service{claude: claude}
}

func (s *Service) GenerateProposal(ctx context.Context, chatHistory []chat.Message, callback chat.StreamCallback) error {
	messages := []chat.Message{
		{Role: "user", Content: formatHistory(chatHistory) + "\n\nPlease generate a proposal document based on our discussion."},
	}
	return s.claude.StreamMessage(ctx, ProposalPrompt, messages, callback)
}

func (s *Service) GenerateSpec(ctx context.Context, chatHistory []chat.Message, callback chat.StreamCallback) error {
	messages := []chat.Message{
		{Role: "user", Content: formatHistory(chatHistory) + "\n\nPlease generate spec requirements based on our discussion."},
	}
	return s.claude.StreamMessage(ctx, SpecPrompt, messages, callback)
}

func (s *Service) SuggestImprovements(ctx context.Context, specContent string, callback chat.StreamCallback) error {
	messages := []chat.Message{
		{Role: "user", Content: "Please review this spec and suggest improvements:\n\n" + specContent},
	}
	return s.claude.StreamMessage(ctx, ImprovementPrompt, messages, callback)
}

func formatHistory(messages []chat.Message) string {
	var sb strings.Builder
	sb.WriteString("Conversation history:\n\n")
	for _, m := range messages {
		sb.WriteString(m.Role + ": " + m.Content + "\n\n")
	}
	return sb.String()
}
