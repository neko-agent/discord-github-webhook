啊明白了！給你一份 PRD（Product Requirements Document）：

```markdown
# GitHub Discord Bridge - Product Requirements Document

## 產品概述

**目標：** 將 GitHub Pull Request 的活動自動同步到 Discord Forum Channel，每個 PR 對應一個獨立的 thread。

**使用情境：** 開發團隊使用 Discord 作為主要溝通工具，希望在 Discord 中即時追蹤 PR 的討論和更新，而不需要頻繁切換到 GitHub。

## 核心功能

### 1. PR 生命週期追蹤

| GitHub 事件                              | Discord 行為                                                                     |
| ---------------------------------------- | -------------------------------------------------------------------------------- |
| PR opened                                | 在 Forum Channel 建立新 thread，標題為 "PR #156: feat(LOVE-77): Add JWT auth..." |
| PR updated (new commits)                 | 在對應 thread 中發送更新通知                                                     |
| Review requested                         | 通知「🔔 @author requested review from @reviewer」（含 re-request）              |
| PR reviewed (approved/changes requested/commented) | 在對應 thread 中顯示 review 結果（含 review type）                      |
| PR merged                                | 發送 merge 訊息，自動 archive thread                                             |
| PR closed (without merge)                | 發送 close 訊息，自動 archive thread                                             |

### 2. 自動補建機制

**問題：** 服務上線前已存在的 PR 沒有對應的 Discord thread

**解決方案：** 當舊 PR 發生新活動（comment/review/update）時，系統自動建立對應的 thread，然後才發送該活動通知

**取捨：**

- ✅ 優點：不需要手動初始化，完全自動化
- ⚠️ 缺點：Thread 建立時間不是 PR 真正開啟的時間，過去的討論不會被補上

### 3. 資料持久化策略

**儲存內容：** PR ID → Discord Thread ID 的對應關係

**生命週期管理：**

- PR 開啟時：儲存 mapping，**不設定過期時間**（永久保存）
- PR 關閉時：設定 **7 天 TTL**
- 7 天後：自動清除（節省儲存空間）

**理由：** PR 關閉後可能還有後續討論，保留 7 天可以應對延遲的 webhook 或補充討論

### 4. 單向同步

**方向：** GitHub → Discord（僅單向）

**Discord 上的討論不會回傳 GitHub**

- 原因：避免無限迴圈、權限問題、訊息污染
- 用途：Discord 用於即時通知和非正式討論，正式 review 仍在 GitHub 進行

## Discord 介面設計

### Thread 結構
```

📋 Code Reviews (Forum Channel)
├── 🟢 PR #156: feat(LOVE-77): Add JWT authentication middleware
│ ├── (Initial Post) - PR 開啟時的詳細資訊
│ ├── 💬 Comment by @sarah - "Should we add unit tests?"
│ ├── ✅ Review by @john - "LGTM! Approved"
│ ├── 🔄 PR Updated - "New commits pushed"
│ └── 🎉 PR Merged - "Merged by @champer-wu"
└── 🟢 PR #155: fix: Memory leak in worker pool

```

### 訊息格式

**PR Opened (Initial Post):**
- 標題：Pull Request #156 Opened
- 內容：PR description（截斷至 500 字）
- 欄位：Author、Branch、Changes (+245 −83)
- 顏色：綠色
- Footer：GitHub icon

**Review (Approved):**
- 標題：✅ Review by @john-reviewer
- 內容：Review comment
- 欄位：Status (✅ Approved)
- 顏色：綠色

**Review (Changes Requested):**
- 標題：🔴 Review by @sarah-dev
- 內容：Review comment
- 欄位：Status (🔴 Changes Requested)
- 顏色：紅色

**Comment:**
- 標題：💬 Comment by @username
- 內容：Comment body
- 顏色：灰色

**PR Merged:**
- 標題：🎉 PR #156 Merged
- 內容："{Title} has been merged into `main`"
- 欄位：Merged by、Changes
- 顏色：紫色
- Footer："Thread will be archived soon"
- 動作：自動 archive thread

## 技術架構

### 系統組件

```

GitHub Webhook
↓
Webhook Handler (驗證簽名 + 解析 payload)
↓
Event Router (根據事件類型分發)
↓
Storage Layer (Redis: PR ID → Thread ID mapping)
↓
Discord Client (建立 thread、發送訊息、archive)

```

### 資料流程

**場景 1：新 PR 開啟**
```

1. GitHub 發送 webhook (action: opened)
2. 系統建立 Discord thread
3. 儲存 "owner/repo#156" → "discord_thread_id" 到 Redis (無 TTL)
4. 在 thread 發送 Initial Post

```

**場景 2：舊 PR 有新 comment**
```

1. GitHub 發送 webhook (action: created, issue_comment)
2. 查詢 Redis: "owner/repo#23" → 不存在
3. 自動建立 thread (呼叫 PR opened 邏輯)
4. 儲存 mapping 到 Redis
5. 在新建的 thread 中發送 comment

```

**場景 3：PR merged**
```

1. GitHub 發送 webhook (action: closed, merged: true)
2. 從 Redis 取得對應的 thread ID
3. 在 thread 發送 merge 訊息
4. 呼叫 Discord API archive thread
5. 更新 Redis mapping，設定 7 天 TTL
6. 7 天後 Redis 自動刪除該 mapping

```

## 環境設定需求

### Discord
- Discord Bot Token（從 Discord Developer Portal 建立）
- Forum Channel ID（目標 forum channel）
- Bot 權限：Send Messages、Create Public Threads、Send Messages in Threads、Manage Threads

### GitHub
- Webhook URL：`https://your-domain.com/webhook/github`
- Webhook Secret（用於驗證請求來源）
- Events：Pull requests、Pull request reviews、Pull request review comments、Issue comments

### Infrastructure
- Redis（儲存 PR-Thread mapping）
- Kubernetes cluster（部署服務）
- 公網 IP 或 LoadBalancer（接收 GitHub webhook）

## 非功能需求

### 可靠性
- Webhook 簽名驗證（防止偽造請求）
- Redis 連線失敗時回傳 500，依賴 GitHub webhook retry 機制

### 效能
- Webhook 處理時間 < 1 秒
- 支援並發處理多個 webhook
- Redis 查詢延遲 < 10ms

### 可維護性
- 結構化 logging（記錄所有事件和錯誤）
- Health check endpoint（`/health`）
- 環境變數配置（不寫死任何 credentials）

## 邊界條件處理

| 情境 | 處理方式 |
|-----|---------|
| PR description 超過 500 字 | 截斷並加上 "..." |
| Thread title 超過 100 字 | 截斷至 97 字 + "..." |
| Discord API 失敗 | Log error，回傳 500 給 GitHub（觸發 retry） |
| Redis 連線失敗 | Log error，回傳 500 給 GitHub（依賴 webhook retry） |
| 收到未知的 webhook event | Log warning 並忽略 |
| 收到 `issue_comment` / `pull_request_review_comment` | Log info 並忽略（不發送通知） |
| `review_requested` 但缺少 `requested_reviewer` | Log warning 並忽略 |
| Payload 中缺少 `pull_request` | Log warning 並忽略（不報錯） |
| 同一個 PR 重複建立 thread | 檢查 Redis 是否已存在，存在則跳過 |

## 成功指標

- **覆蓋率：** 100% 的 PR 活動都能同步到 Discord
- **延遲：** Webhook 到 Discord 通知的時間 < 3 秒
- **可靠性：** 99.9% 的 webhook 成功處理
- **使用者滿意度：** 團隊減少 50% 切換到 GitHub 查看 PR 的次數

## TODO（測試後修正）

### 事件路由重構
- [x] 使用 `X-GitHub-Event` header 做事件路由（取代 `payload.EventType()` 的純 payload 內容判斷）
- [x] 移除 `EventType()` method（不再需要）

### 移除 comment 通知
- [x] 忽略 `issue_comment` 和 `pull_request_review_comment` 事件（header 路由直接跳過）
- [x] 移除 `handlePRCommented()`、`FormatPRComment()`
- [x] 移除 `WebhookPayload.Comment` 欄位（不再使用）

### 新增 review_requested 通知
- [x] `WebhookPayload` 新增 `RequestedReviewer` 欄位
- [x] 新增 `handleReviewRequested()`（含 re-request review）
- [x] 新增 `FormatReviewRequested()`（通知「🔔 @author requested review from @reviewer」）

### 強化 review 通知
- [x] 確認 `pull_request_review`（submitted）通知包含 review type（approved / changes_requested / commented）

### PRD 更新
- [x] 更新 PR 生命週期追蹤表格（移除 PR commented，新增 review requested）
- [x] 更新邊界條件處理表格

### CI/CD 狀態通知（改用 `workflow_run`，因為 GitHub Actions 不觸發 `check_suite`）
- [ ] GitHub webhook 設定加上 `Workflow runs` 事件
- [x] `WebhookPayload` 新增 `WorkflowRun` struct（`name`、`conclusion`、`head_sha`、`html_url`、`pull_requests`）
- [x] 路由 `X-GitHub-Event: workflow_run`，只處理 `action: "completed"`
- [x] 新增 `handleWorkflowRunCompleted()`（關聯 PR thread，只發送 success/failure 通知）
- [x] 新增 `FormatWorkflowRunResult()`（顯示 workflow 名稱、commit、連結到 GitHub run）

## 未來擴展

- [ ] 支援多個 repository 對應不同 Forum Channel
- [ ] 支援 Forum Tags（用於分類 PR：bug、feature、urgent）
- [ ] 統計 Dashboard（PR 平均 review 時間、活躍度）
- [ ] Discord API rate limit 處理（當支援多 repo / 高頻率事件時）
- [ ] 結構化 logging 接入 centralized logging（ELK、Datadog 等）
- [ ] GitHub User ↔ Discord User 對應（支援 @mention Discord 用戶，讓通知更有感）
```

Reference
Github webhook Official Doc: https://docs.github.com/en/webhooks/webhook-events-and-payloads
