# Structured Proposal — 설계 결정 문서

## 배경

Colign의 Change는 `Draft → Design → Review → Ready` 워크플로우를 따른다.
기존 Proposal 탭은 TipTap 기반 자유 텍스트 에디터였다. AI 시대에 SDD(Spec-Driven Development)가 다시 떠오르면서, 스펙의 구조화가 필수가 되었다.

### 왜 구조화가 필요한가

1. **스펙 = AI의 컨텍스트** — AI 에이전트가 코드를 생성할 때, 구조화된 스펙이 정확도를 결정한다 (Red Hat: SDD 적용 시 AI 생성 코드 정확도 95%+)
2. **플랫폼 내 AI + 외부 AI 연동** — MCP Server를 통해 Claude Code/Cursor 등이 스펙을 읽고 쓸 때, JSON 구조가 필수
3. **직군 간 협업** — PM/디자이너/엔지니어가 각자의 관점에서 기여할 명확한 영역이 필요

## 경쟁 제품 비교 후 결정

6개 SDD 도구의 스펙 구조를 조사했다:

| 도구 | 스펙 구조 | 특징 |
|------|----------|------|
| **GitHub Spec Kit** | User Scenarios + Requirements + Success Criteria | `[NEEDS CLARIFICATION]` 마커 |
| **Amazon Kiro** | requirements.md (EARS) + design.md + tasks.md | EARS 구문 (WHEN/SHALL) |
| **OpenSpec** | Intent + Scope + Approach → Delta Specs | ADDED/MODIFIED/REMOVED 변경분 |
| **Superpowers** | Socratic 브레인스토밍 → Design Doc → 2-5분 태스크 | Socratic 질문 기반 |
| **Tessl** | Functional Requirements + API Contracts + @test links | 스펙이 코드의 원본 |
| **Addy Osmani** | Objective + Tech Stack + Boundaries + Success Criteria | Always/Ask/Never 경계 |

### 공통점 (전원 필수)

- WHY — 왜 하는가
- WHAT — 뭘 하는가
- Acceptance Criteria / Scenarios
- Tasks

### Colign의 차별점

- **Project Memory** — 다른 도구에 없는 프로젝트 컨텍스트 자동 주입
- **AC가 이미 별도 서비스** — Given/When/Then 구조로 독립 구현 완료
- **실시간 협업** — Y.js + Hocuspocus 기반 동시 편집

## 최종 섹션 구조

기존 AC 서비스를 제외하면, Proposal은 4개 섹션으로 구성:

```
§ Problem         (필수)  — 왜 이 변경이 필요한가
§ Scope           (필수)  — 무엇이 변경되는가
§ Out of Scope    (선택)  — 명시적으로 하지 않는 것
§ Approach        (선택)  — 기술적 방향과 근거

+ Acceptance Criteria (별도 컴포넌트, 이미 존재)
+ Project Memory (프로젝트 레벨, 자동 참조)
```

### 왜 4개인가

- **최소주의**: 다른 도구가 6-8개 섹션인데, 핵심만 남김. 무거운 스펙은 작성률을 떨어뜨린다.
- **AI 친화적**: 4개 필드의 JSON은 어떤 AI 도구에서든 파싱/생성이 쉽다.
- **점진적 노출**: 필수 2개(Problem, Scope)만 항상 보이고, 선택 2개는 접혀있음.

### 왜 이 섹션들인가

| 섹션 | 근거 | 출처 |
|------|------|------|
| Problem | 모든 도구가 "왜"를 첫 섹션으로 둠 | 전원 공통 |
| Scope | OpenSpec의 Intent+Scope 패턴, Spec Kit의 Requirements | OpenSpec, Spec Kit |
| Out of Scope | Kiro design.md의 Non-Goals, 범위 관리의 핵심 | Kiro, Osmani |
| Approach | OpenSpec의 Approach, Kiro design.md의 Architecture | OpenSpec, Kiro |

## 컴포넌트 아키텍처

```
Change Detail Page
├── Stage Stepper (Draft → Design → Review → Ready)
├── Gate Conditions
├── Tab Navigation
│   ├── Proposal ──→ StructuredProposal (신규)
│   │                 ├── SectionEditor × 4 (Problem, Scope, Out of Scope, Approach)
│   │                 ├── AI Generate (coming soon)
│   │                 └── AcceptanceCriteria (기존 컴포넌트)
│   ├── Design ────→ DocumentTab + SpecEditor (TipTap, 기존)
│   ├── Specs ─────→ DocumentTab + SpecEditor (TipTap, 기존)
│   ├── Tasks ─────→ TaskBoard (기존)
│   └── History ───→ WorkflowEvent list (기존)
```

### 파일 구조

```
web/src/components/change/
├── structured-proposal.tsx   ← 신규: 구조화 Proposal
├── acceptance-criteria.tsx   ← 기존: AC (Given/When/Then)
├── document-tab.tsx          ← 기존: Design/Specs 탭용 (자유 텍스트)

web/src/components/editor/
├── spec-editor.tsx           ← 기존: TipTap + Y.js 에디터
├── templates.ts              ← 기존: Design/Specs 템플릿
```

### 저장 방식

현재(Phase 1): 기존 `Document.content`에 JSON 문자열로 저장.

```json
{
  "problem": "모바일 결제에서 Apple Pay 미지원으로 전환율 15% 저하",
  "scope": "- Stripe Apple Pay 결제 연동\n- iOS Safari에서 결제 버튼 표시\n- 결제 완료 콜백 처리",
  "outOfScope": "- Android Pay\n- 직접 PG 연동",
  "approach": "Stripe의 Payment Request API를 사용. 기존 checkout flow에 분기 추가."
}
```

미래(Phase 2): 별도 `StructuredSpec` proto 서비스로 분리, MCP Server에서 직접 섹션별 CRUD.

### 레거시 호환

기존에 자유 텍스트(HTML)로 작성된 Proposal이 있을 경우, `parseContent()` 함수가:
1. JSON 파싱 시도 → 성공하면 구조화 데이터
2. 실패하면 HTML 태그 제거 후 `problem` 섹션에 전체 텍스트 삽입

## 향후 계획

| Phase | 내용 |
|-------|------|
| Phase 1 (현재) | 프론트엔드 구조화, 기존 Document API에 JSON 저장 |
| Phase 2 | `StructuredSpec` proto 서비스 추가, 섹션별 CRUD API |
| Phase 3 | 플랫폼 내 AI 생성 ("Generate" 버튼 → Claude API → 구조화 초안) |
| Phase 4 | MCP Server — 외부 AI 도구에서 스펙 읽기/쓰기 |
| Phase 5 | Role-aware Review — PM/Engineer/Designer 관점별 리뷰 체크리스트 |
