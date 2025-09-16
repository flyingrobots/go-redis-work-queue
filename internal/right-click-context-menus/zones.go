package context_menus

import (
	"fmt"
)

// NewZoneManager creates a new zone manager
func NewZoneManager() *ZoneManager {
	return &ZoneManager{
		zones:   make(map[string]BubbleZone),
		enabled: true,
	}
}

// RegisterZone registers a clickable zone
func (zm *ZoneManager) RegisterZone(zone BubbleZone) error {
	if zone.ID == "" {
		return &ContextMenuError{
			Message: "zone ID cannot be empty",
			Code:    "EMPTY_ZONE_ID",
		}
	}

	zm.zones[zone.ID] = zone
	return nil
}

// UnregisterZone removes a zone from the manager
func (zm *ZoneManager) UnregisterZone(id string) {
	delete(zm.zones, id)
}

// GetZoneAt returns the topmost zone at the given coordinates
func (zm *ZoneManager) GetZoneAt(x, y int) (BubbleZone, bool) {
	if !zm.enabled {
		return BubbleZone{}, false
	}

	// Check zones in reverse order (last registered = topmost)
	for _, zone := range zm.zones {
		if !zone.Enabled {
			continue
		}

		if x >= zone.X && x < zone.X+zone.Width &&
			y >= zone.Y && y < zone.Y+zone.Height {
			return zone, true
		}
	}

	return BubbleZone{}, false
}

// GetZone returns a zone by ID
func (zm *ZoneManager) GetZone(id string) (BubbleZone, bool) {
	zone, exists := zm.zones[id]
	return zone, exists
}

// ListZones returns all registered zones
func (zm *ZoneManager) ListZones() []BubbleZone {
	zones := make([]BubbleZone, 0, len(zm.zones))
	for _, zone := range zm.zones {
		zones = append(zones, zone)
	}
	return zones
}

// ClearZones removes all zones
func (zm *ZoneManager) ClearZones() {
	zm.zones = make(map[string]BubbleZone)
}

// SetEnabled enables or disables zone detection
func (zm *ZoneManager) SetEnabled(enabled bool) {
	zm.enabled = enabled
}

// IsEnabled returns whether zone detection is enabled
func (zm *ZoneManager) IsEnabled() bool {
	return zm.enabled
}

// UpdateZone updates an existing zone's properties
func (zm *ZoneManager) UpdateZone(id string, updater func(*BubbleZone)) error {
	zone, exists := zm.zones[id]
	if !exists {
		return ErrZoneNotFound
	}

	updater(&zone)
	zm.zones[id] = zone
	return nil
}

// RegisterTableRow registers a zone for a table row
func (zm *ZoneManager) RegisterTableRow(rowIndex int, x, y, width int, queueName string) error {
	zoneID := fmt.Sprintf("table-row-%d", rowIndex)
	zone := BubbleZone{
		ID:     zoneID,
		X:      x,
		Y:      y,
		Width:  width,
		Height: 1, // Single row height
		Context: MenuContext{
			Type:      ContextQueueRow,
			QueueName: queueName,
			RowIndex:  rowIndex,
			Position:  Position{X: x, Y: y},
			Metadata: map[string]interface{}{
				"rowIndex":  rowIndex,
				"queueName": queueName,
			},
		},
		Enabled: true,
	}

	return zm.RegisterZone(zone)
}

// RegisterTab registers a zone for a tab
func (zm *ZoneManager) RegisterTab(tabID, label string, x, y, width int) error {
	zoneID := fmt.Sprintf("tab-%s", tabID)
	zone := BubbleZone{
		ID:     zoneID,
		X:      x,
		Y:      y,
		Width:  width,
		Height: 1, // Single line height
		Context: MenuContext{
			Type:     ContextTab,
			ItemID:   tabID,
			Position: Position{X: x, Y: y},
			Metadata: map[string]interface{}{
				"tabID": tabID,
				"label": label,
			},
		},
		Enabled: true,
	}

	return zm.RegisterZone(zone)
}

// RegisterChart registers a zone for a chart area
func (zm *ZoneManager) RegisterChart(chartID string, x, y, width, height int) error {
	zoneID := fmt.Sprintf("chart-%s", chartID)
	zone := BubbleZone{
		ID:     zoneID,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Context: MenuContext{
			Type:     ContextChart,
			ItemID:   chartID,
			Position: Position{X: x, Y: y},
			Metadata: map[string]interface{}{
				"chartID": chartID,
			},
		},
		Enabled: true,
	}

	return zm.RegisterZone(zone)
}

// RegisterInfoRegion registers a zone for an info region
func (zm *ZoneManager) RegisterInfoRegion(regionID string, x, y, width, height int) error {
	zoneID := fmt.Sprintf("info-%s", regionID)
	zone := BubbleZone{
		ID:     zoneID,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Context: MenuContext{
			Type:     ContextInfoRegion,
			ItemID:   regionID,
			Position: Position{X: x, Y: y},
			Metadata: map[string]interface{}{
				"regionID": regionID,
			},
		},
		Enabled: true,
	}

	return zm.RegisterZone(zone)
}

// RegisterDLQItem registers a zone for a DLQ item
func (zm *ZoneManager) RegisterDLQItem(itemIndex int, jobID string, x, y, width int) error {
	zoneID := fmt.Sprintf("dlq-item-%d", itemIndex)
	zone := BubbleZone{
		ID:     zoneID,
		X:      x,
		Y:      y,
		Width:  width,
		Height: 1, // Single row height
		Context: MenuContext{
			Type:     ContextDLQItem,
			ItemID:   jobID,
			JobID:    jobID,
			RowIndex: itemIndex,
			Position: Position{X: x, Y: y},
			Metadata: map[string]interface{}{
				"itemIndex": itemIndex,
				"jobID":     jobID,
			},
		},
		Enabled: true,
	}

	return zm.RegisterZone(zone)
}