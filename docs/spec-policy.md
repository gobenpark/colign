# Colign 스펙 정책

## 스펙 상태 (Lifecycle)

```
Draft → In Review → Approved → Dispatched
                                    ↓
                 (어디서든) → Archived
```

### 각 상태 정의

| 상태 | 설명 |
|------|------|
| **Draft** | 작성/편집 중. 자유롭게 편집 가능. 실시간 공동 편집 가능. 아직 비공식. |
| **In Review** | 공식 리뷰 요청됨. 본문 편집 잠김. Suggest(제안)와 코멘트만 가능. 리뷰어가 읽고 판단하는 단계. |
| **Approved** | Approver 전원 승인 완료. 확정된 스펙. 직접 편집 불가. 수정하려면 반드시 Amend를 통해 새 버전 생성. |
| **Dispatched** | 하류 도구(Linear, Claude Code, Cursor, Paperclip 등)에 전달 완료. 어디로 갔는지 추적됨. 편집 불가. 수정하려면 Amend. |
| **Archived** | 더 이상 활성 아님. 검색/참조용으로 보존. 통합되었거나 폐기된 스펙. |

## 상태별 권한 매트릭스

| | 본문 편집 | Suggest | 코멘트 | Amend |
|---|---|---|---|---|
| **Draft** | ✅ 자유 | ✅ | ✅ | — |
| **In Review** | 🔒 잠김 | ✅ | ✅ | — |
| **Approved** | 🔒 | 🔒 | ✅ | ✅ 새 버전 생성 |
| **Dispatched** | 🔒 | 🔒 | ✅ | ✅ 새 버전 생성 |
| **Archived** | 🔒 | 🔒 | 🔒 | — |

## 상태 전이 규칙

- **Draft → In Review**: 작성자가 리뷰 요청. Approver 최소 1명 지정 필수.
- **In Review → Draft**: 작성자가 리뷰 철회하고 다시 편집. 또는 Changes Requested 받은 후 수정하려고 돌리는 경우. 리뷰 기록은 보존.
- **In Review → Approved**: Approver 전원 Approve 시 자동 전환.
- **Approved → Dispatched**: 디스패치 실행 시.
- **Approved/Dispatched → Amend**: 새 버전(v2) Draft 생성. 원본은 보존.
- **어디서든 → Archived**: 수동.

## 리뷰 역할 (3종)

### 결재자 (Approver)
승인 권한 있음. Changes Requested는 **블로킹** — 이 사람이 Approve 안 하면 스펙이 안 넘어감. 전원 승인 필수. 리뷰 요청 시 최소 1명 지정 필수.

### 검토자 (Reviewer)
의견 남길 수 있고 Changes Requested도 가능하지만 **논블로킹** — Approver가 다 승인하면 Reviewer의 반대와 무관하게 Approved 전환 가능. 다만 Reviewer의 미해결 코멘트가 있으면 Approver에게 경고 표시.

### 참조 (CC)
알림만 받고 읽기만 함. 코멘트는 가능. 승인/거절 권한 없음.

## Amendment (수정) 정책

Approved 또는 Dispatched 된 스펙을 직접 수정하지 않는다. 수정이 필요하면 Amend를 통해 새 버전 Draft를 생성하고, 동일한 리뷰 사이클을 탄다.

```
Approved v1 → Amend 클릭 → Draft v2 → In Review v2 → Approved v2
```

- 원본 v1은 히스토리에 보존.
- v1과 v2 사이 diff 언제든 확인 가능.
- Dispatched 된 버전이 v1이면 v2 Approved 후 재디스패치 가능.
- "지금 배포된 건 v1, 수정 중인 건 v2" 상태가 한눈에 보임.

**핵심 규칙**: Amend는 한 번에 하나만. 같은 스펙에서 v2-a, v2-b가 동시에 열리지 않는다. 누가 먼저 Amend를 열면, 다른 사람은 그 Draft v2에 공동 편집으로 참여. 각자 따로 쓰고 나중에 합치는 건 병목이니까 — 같이 써라.

## 스펙 통합 정책

서로 다른 스펙 두 개를 하나로 합쳐야 하는 경우 (예: "결제 스펙"과 "구독 스펙"이 사실 하나여야 했을 때):

1. 새 스펙을 만든다.
2. "기존 스펙에서 가져오기"로 두 스펙의 내용을 가져와서 편집한다.
3. 기존 두 스펙은 Archived 처리하며 "→ [새 스펙]으로 통합됨" 링크가 자동으로 남는다.

> 나중에 시스템이 성숙하면 섹션 단위 선택 머지 UI를 지원할 수 있지만, MVP에서는 위 방식으로 충분.

## 우선순위

스펙마다 4단계 우선순위:

| 우선순위 | 설명 |
|----------|------|
| **Urgent** | 즉시 처리. 리뷰/승인 지연 시 블로커 알림. |
| **High** | 빠르게 처리. |
| **Medium** | 일반. |
| **Low** | 여유 있을 때. |

## 병목 감지 (Human is the Bottleneck)

- **3일 이상** 같은 상태에 머문 스펙은 병목으로 표시.
- 누가 블로커인지(Approver가 승인 안 함, 작성자가 수정 안 함 등) 명시.
- **"내가 블로커인 스펙"**은 Home 최상단에 빨간 박스로 강조.
- Approved인데 Dispatched 안 된 스펙도 **"미디스패치"**로 표시 — 확정됐는데 아무도 안 넘기고 있으면 그것도 병목.
- Amend 진행 중인 스펙(v1 Dispatched + v2 Draft/Review)은 **"진실이 갈라진 상태"**로 별도 표시.
