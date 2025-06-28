package builders

import (
	"pvz-cli/internal/common/clock"
	"pvz-cli/internal/models"
	"time"
)

type HistoryEntryBuilder struct {
	clk   clock.Clock
	entry models.HistoryEntry
}

// NewHistoryEntryBuilder creates a new HistoryEntryBuilder initialized with the current timestamp from the provided clock.
func NewHistoryEntryBuilder(clk clock.Clock) *HistoryEntryBuilder {
	return &HistoryEntryBuilder{
		clk: clk,
		entry: models.HistoryEntry{
			Timestamp: clk.Now(),
		},
	}
}

func (b *HistoryEntryBuilder) WithOrderID(id uint64) *HistoryEntryBuilder {
	b.entry.OrderID = id
	return b
}

func (b *HistoryEntryBuilder) WithEvent(event models.EventType) *HistoryEntryBuilder {
	b.entry.Event = event
	return b
}

func (b *HistoryEntryBuilder) WithOffset(offset time.Duration) *HistoryEntryBuilder {
	b.entry.Timestamp = b.entry.Timestamp.Add(offset)
	return b
}

func (b *HistoryEntryBuilder) Build() models.HistoryEntry {
	return b.entry
}
