package storage

// Store 定義 PR → Discord Thread ID 的儲存介面
type Store interface {
	// Set 儲存 PR 和 Thread 的對應關係（無 TTL）
	Set(prID, threadID string) error

	// Get 取得對應的 Thread ID
	Get(prID string) (threadID string, exists bool, err error)

	// Delete 刪除對應關係（少用，通常用 MarkAsClosed）
	Delete(prID string) error

	// MarkAsClosed 標記 PR 已關閉，設定 7 天 TTL
	MarkAsClosed(prID string) error
}
