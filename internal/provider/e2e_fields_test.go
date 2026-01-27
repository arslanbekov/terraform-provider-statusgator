package provider

import (
	"context"
	"os"
	"testing"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_MonitorFields tests that all monitor fields are correctly mapped
func TestE2E_MonitorFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	// Test Monitors.List - this is what Read operations use
	monitors, pagination, err := client.Monitors.List(ctx, boardID, nil)
	require.NoError(t, err)
	require.NotNil(t, pagination)

	t.Logf("Found %d monitors", len(monitors))

	for _, m := range monitors {
		t.Logf("Monitor: ID=%s, DisplayName=%s, MonitorType=%s, FilteredStatus=%s",
			m.ID, m.DisplayName, m.MonitorType, m.FilteredStatus)

		// Verify required fields exist and are populated
		assert.NotEmpty(t, m.ID, "ID should not be empty")
		assert.NotEmpty(t, m.DisplayName, "DisplayName should not be empty")
		assert.NotEmpty(t, m.MonitorType, "MonitorType should not be empty")
		assert.NotEmpty(t, m.FilteredStatus, "FilteredStatus should not be empty")
		assert.NotEmpty(t, m.IconURL, "IconURL should not be empty")

		// Verify MonitorType is valid
		validTypes := []statusgator.MonitorType{
			statusgator.MonitorTypeWebsite,
			statusgator.MonitorTypePing,
			statusgator.MonitorTypeService,
			statusgator.MonitorTypeCustom,
		}
		assert.Contains(t, validTypes, m.MonitorType)

		// Verify FilteredStatus is valid
		validStatuses := []statusgator.MonitorStatus{
			statusgator.MonitorStatusUp,
			statusgator.MonitorStatusDown,
			statusgator.MonitorStatusWarn,
			statusgator.MonitorStatusMaintenance,
			statusgator.MonitorStatusUnknown,
		}
		assert.Contains(t, validStatuses, m.FilteredStatus)

		// Test Group field if present
		if m.Group != nil {
			t.Logf("  Group: ID=%s, Name=%s", m.Group.ID, m.Group.Name)
			assert.NotEmpty(t, m.Group.ID, "Group.ID should not be empty when Group is set")
			assert.NotEmpty(t, m.Group.Name, "Group.Name should not be empty when Group is set")
		}

		// Test Service field if present (for ServiceMonitors)
		if m.Service != nil {
			t.Logf("  Service: ID=%s, Name=%s, Slug=%s", m.Service.ID, m.Service.Name, m.Service.Slug)
			assert.NotEmpty(t, m.Service.ID, "Service.ID should not be empty")
			assert.NotEmpty(t, m.Service.Name, "Service.Name should not be empty")
		}

		// Test IsPaused helper
		_ = m.IsPaused() // Should not panic
	}
}

// TestE2E_BoardFields tests board fields
func TestE2E_BoardFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	boards, pagination, err := client.Boards.List(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, pagination)
	require.NotEmpty(t, boards)

	for _, b := range boards {
		t.Logf("Board: ID=%s, Name=%s, PublicToken=%s", b.ID, b.Name, b.PublicToken)

		assert.NotEmpty(t, b.ID, "ID should not be empty")
		assert.NotEmpty(t, b.Name, "Name should not be empty")
		assert.NotEmpty(t, b.PublicToken, "PublicToken should not be empty")
		assert.NotZero(t, b.CreatedAt, "CreatedAt should not be zero")
	}
}

// TestE2E_MonitorGroupFields tests monitor group fields
func TestE2E_MonitorGroupFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	groups, err := client.MonitorGroups.List(ctx, boardID)
	require.NoError(t, err)

	t.Logf("Found %d monitor groups", len(groups))

	for _, g := range groups {
		t.Logf("MonitorGroup: ID=%s, Name=%s, Position=%d", g.ID, g.Name, g.Position)

		assert.NotEmpty(t, g.ID, "ID should not be empty")
		assert.NotEmpty(t, g.Name, "Name should not be empty")
		assert.NotZero(t, g.CreatedAt, "CreatedAt should not be zero")
	}
}

// TestE2E_IncidentFields tests incident fields
func TestE2E_IncidentFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	incidents, pagination, err := client.Incidents.List(ctx, boardID, nil)
	require.NoError(t, err)
	require.NotNil(t, pagination)

	t.Logf("Found %d incidents", len(incidents))

	for _, inc := range incidents {
		t.Logf("Incident: ID=%s, Name=%s, Severity=%s, Phase=%s, BoardID=%s",
			inc.ID, inc.Name, inc.Severity, inc.Phase, inc.BoardID)

		assert.NotEmpty(t, inc.ID, "ID should not be empty")
		assert.NotEmpty(t, inc.Name, "Name should not be empty")
		assert.NotEmpty(t, inc.Severity, "Severity should not be empty")
		assert.NotEmpty(t, inc.Phase, "Phase should not be empty")
		assert.NotEmpty(t, inc.BoardID, "BoardID should not be empty")
		assert.NotZero(t, inc.CreatedAt, "CreatedAt should not be zero")

		// Details might be empty for some incidents
		t.Logf("  Details: %s", inc.Details)

		// Check scheduled maintenance fields
		if inc.ScheduledMaintenance {
			t.Logf("  Scheduled maintenance: WillStartAt=%v, WillEndAt=%v", inc.WillStartAt, inc.WillEndAt)
		}
	}
}

// TestE2E_RegionsFields tests regions fields
func TestE2E_RegionsFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	regions, err := client.Regions.List(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, regions)

	t.Logf("Found %d regions", len(regions))

	for _, r := range regions {
		t.Logf("Region: ID=%s, Name=%s, Code=%s, Provider=%s", r.RegionID, r.Name, r.Code, r.Provider)

		assert.NotEmpty(t, r.RegionID, "RegionID should not be empty")
		assert.NotEmpty(t, r.Name, "Name should not be empty")
		assert.NotEmpty(t, r.Code, "Code should not be empty")
		assert.NotEmpty(t, r.Provider, "Provider should not be empty")
	}
}

// TestE2E_Ping tests API connectivity
func TestE2E_Ping(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed")
}

// TestE2E_WebsiteMonitorFields tests WebsiteMonitor fields from List
func TestE2E_WebsiteMonitorFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	// Get all monitors
	monitors, _, err := client.Monitors.List(ctx, boardID, nil)
	require.NoError(t, err)

	// Find WebsiteMonitors and verify fields
	var found bool
	for _, m := range monitors {
		if m.MonitorType == statusgator.MonitorTypeWebsite {
			found = true
			t.Logf("WebsiteMonitor from List: ID=%s, DisplayName=%s, FilteredStatus=%s",
				m.ID, m.DisplayName, m.FilteredStatus)

			// Verify fields used in provider Read
			assert.NotEmpty(t, m.ID, "ID should not be empty")
			assert.NotEmpty(t, m.DisplayName, "DisplayName should not be empty")
			assert.NotEmpty(t, m.FilteredStatus, "FilteredStatus should not be empty")
			assert.Equal(t, statusgator.MonitorTypeWebsite, m.MonitorType)

			// IsPaused helper
			_ = m.IsPaused()

			// Group field
			if m.Group != nil {
				t.Logf("  Group: ID=%s, Name=%s", m.Group.ID, m.Group.Name)
				assert.NotEmpty(t, m.Group.ID)
			}
		}
	}

	if !found {
		t.Skip("No WebsiteMonitor found in board")
	}
}

// TestE2E_ServiceMonitorFields tests ServiceMonitor fields from List
func TestE2E_ServiceMonitorFields(t *testing.T) {
	token := os.Getenv("STATUSGATOR_API_TOKEN")
	if token == "" {
		t.Skip("STATUSGATOR_API_TOKEN not set")
	}

	boardID := os.Getenv("STATUSGATOR_TEST_BOARD_ID")
	if boardID == "" {
		t.Skip("STATUSGATOR_TEST_BOARD_ID not set")
	}

	client, err := statusgator.NewClient(token)
	require.NoError(t, err)

	ctx := context.Background()

	// Get all monitors
	monitors, _, err := client.Monitors.List(ctx, boardID, nil)
	require.NoError(t, err)

	// Find ServiceMonitors and verify fields
	var found bool
	for _, m := range monitors {
		if m.MonitorType == statusgator.MonitorTypeService {
			found = true
			t.Logf("ServiceMonitor from List: ID=%s, DisplayName=%s, FilteredStatus=%s",
				m.ID, m.DisplayName, m.FilteredStatus)

			// Verify fields used in provider Read
			assert.NotEmpty(t, m.ID, "ID should not be empty")
			assert.NotEmpty(t, m.DisplayName, "DisplayName should not be empty")
			assert.NotEmpty(t, m.FilteredStatus, "FilteredStatus should not be empty")
			assert.Equal(t, statusgator.MonitorTypeService, m.MonitorType)

			// Service field should be present for ServiceMonitors
			if m.Service != nil {
				t.Logf("  Service: ID=%s, Name=%s, Slug=%s", m.Service.ID, m.Service.Name, m.Service.Slug)
				assert.NotEmpty(t, m.Service.ID, "Service.ID should not be empty")
				assert.NotEmpty(t, m.Service.Name, "Service.Name should not be empty")
			}

			// Group field
			if m.Group != nil {
				t.Logf("  Group: ID=%s, Name=%s", m.Group.ID, m.Group.Name)
				assert.NotEmpty(t, m.Group.ID)
			}
		}
	}

	if !found {
		t.Skip("No ServiceMonitor found in board")
	}
}
