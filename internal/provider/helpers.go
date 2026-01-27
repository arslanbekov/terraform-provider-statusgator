package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// importBoardChildState is a shared helper for importing resources that belong to a board.
// Import ID format: board_id/resource_id
func importBoardChildState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: board_id/resource_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("board_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// urlRegex returns a compiled regex for validating HTTP/HTTPS URLs.
func urlRegex() *regexp.Regexp {
	return regexp.MustCompile(`^https?://[^\s/$.?#]\.[^\s]*$`)
}

// emailRegex returns a compiled regex for validating email addresses.
func emailRegex() *regexp.Regexp {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
}
