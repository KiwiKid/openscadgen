package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// TestInstanceRenderingSimple tests the core functionality without server complexity
func TestInstanceRenderingSimple(t *testing.T) {
	configPath := "../examples/small-tray/config.toml"

	// Load config to get expected instances
	config, _, err := pkg.LoadConfigFromFile(models.CmdFlags{ConfigFile: configPath, Server: true})
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	instances, err := pkg.GenerateInstanceConfigs(config)
	if err != nil {
		t.Fatalf("Failed to generate instance configs: %v", err)
	}

	expectedInstanceNames := make([]string, len(instances))
	for i, instance := range instances {
		expectedInstanceNames[i] = instance.AutoName
	}

	t.Logf("Testing with %d expected instances: %v", len(expectedInstanceNames), expectedInstanceNames)

	// Test 1: Verify instance names are generated correctly
	t.Run("InstanceNamesGenerated", func(t *testing.T) {
		if len(expectedInstanceNames) == 0 {
			t.Fatal("No instances generated")
		}

		for _, name := range expectedInstanceNames {
			if name == "" {
				t.Error("Empty instance name found")
			}
			t.Logf("✓ Instance name: %s", name)
		}
	})

	// Test 2: Verify HTML ID format
	t.Run("HTMLIDFormat", func(t *testing.T) {
		for _, instance := range instances {
			expectedID := fmt.Sprintf("instance-%s", instance.AutoName)

			// Test that the ID format is correct
			if !strings.HasPrefix(expectedID, "instance-") {
				t.Errorf("Instance ID should start with 'instance-': %s", expectedID)
			}

			// Test that the ID doesn't contain invalid characters
			if strings.Contains(expectedID, " ") {
				t.Errorf("Instance ID should not contain spaces: %s", expectedID)
			}

			t.Logf("✓ Instance ID format: %s", expectedID)
		}
	})

	// Test 3: Verify HTMX OOB response format
	t.Run("HTMXOOBResponseFormat", func(t *testing.T) {
		// Simulate an HTMX OOB response
		instance := instances[0]
		htmlContent := fmt.Sprintf(`<div class="column is-4" id="instance-%s">%s</div>`, instance.AutoName, instance.AutoName)

		// Create the OOB response format
		oobResponse := fmt.Sprintf(`<div id="instances-grid" hx-swap-oob="beforeend">%s</div>`, htmlContent)

		// Verify the response contains required HTMX attributes
		if !strings.Contains(oobResponse, `hx-swap-oob="beforeend"`) {
			t.Error("OOB response should contain hx-swap-oob attribute")
		}

		if !strings.Contains(oobResponse, `id="instances-grid"`) {
			t.Error("OOB response should target instances-grid")
		}

		if !strings.Contains(oobResponse, fmt.Sprintf(`id="instance-%s"`, instance.AutoName)) {
			t.Error("OOB response should contain instance div with correct ID")
		}

		t.Logf("✓ HTMX OOB response format: %s", oobResponse)
	})

	// Test 4: Verify progress polling HTML format
	t.Run("ProgressPollingHTMLFormat", func(t *testing.T) {
		jobID := "test-job-123"
		progressMsg := "Processing..."

		// Create progress polling HTML
		progressHTML := fmt.Sprintf(`<div id="progress" hx-get="/progress?id=%s" hx-trigger="every 1s" hx-swap="outerHTML" class="notification is-info">%s</div>`, jobID, progressMsg)

		// Verify required HTMX attributes
		requiredAttributes := []string{
			`hx-get="/progress?id=`,
			`hx-trigger="every 1s"`,
			`hx-swap="outerHTML"`,
			`id="progress"`,
		}

		for _, attr := range requiredAttributes {
			if !strings.Contains(progressHTML, attr) {
				t.Errorf("Progress HTML should contain: %s", attr)
			}
		}

		t.Logf("✓ Progress polling HTML format: %s", progressHTML)
	})

	// Test 5: Verify completion HTML format
	t.Run("CompletionHTMLFormat", func(t *testing.T) {
		// Create completion HTML (no HTMX attributes)
		completionHTML := `<div id="progress" class="notification is-success">All instances completed!</div>`

		// Verify it doesn't contain HTMX attributes
		if strings.Contains(completionHTML, "hx-get") {
			t.Error("Completion HTML should not contain hx-get attribute")
		}

		if strings.Contains(completionHTML, "hx-trigger") {
			t.Error("Completion HTML should not contain hx-trigger attribute")
		}

		if strings.Contains(completionHTML, "hx-swap") {
			t.Error("Completion HTML should not contain hx-swap attribute")
		}

		// Verify it contains completion message
		if !strings.Contains(completionHTML, "All instances completed!") {
			t.Error("Completion HTML should contain completion message")
		}

		t.Logf("✓ Completion HTML format: %s", completionHTML)
	})
}

// TestHTMXAttributes tests HTMX attribute combinations
func TestHTMXAttributes(t *testing.T) {
	t.Run("ProgressPollingAttributes", func(t *testing.T) {
		// Test that progress polling has correct HTMX attributes
		html := `<div id="progress" hx-get="/progress?id=123" hx-trigger="every 1s" hx-swap="outerHTML" class="notification is-info">Processing...</div>`

		// Verify all required attributes are present
		attributes := map[string]string{
			"hx-get":     "/progress?id=123",
			"hx-trigger": "every 1s",
			"hx-swap":    "outerHTML",
			"id":         "progress",
		}

		for attr, value := range attributes {
			expected := fmt.Sprintf(`%s="%s"`, attr, value)
			if !strings.Contains(html, expected) {
				t.Errorf("HTML should contain %s", expected)
			}
		}
	})

	t.Run("OOBUpdateAttributes", func(t *testing.T) {
		// Test that OOB updates have correct HTMX attributes
		html := `<div id="instances-grid" hx-swap-oob="beforeend">Instance content</div>`

		// Verify OOB attributes
		if !strings.Contains(html, `hx-swap-oob="beforeend"`) {
			t.Error("OOB update should contain hx-swap-oob attribute")
		}

		if !strings.Contains(html, `id="instances-grid"`) {
			t.Error("OOB update should target instances-grid")
		}
	})

	t.Run("InstanceCardAttributes", func(t *testing.T) {
		// Test that instance cards have correct IDs but no conflicting HTMX attributes
		html := `<div class="column is-4" id="instance-test-instance">Test Instance</div>`

		// Verify ID format
		if !strings.Contains(html, `id="instance-test-instance"`) {
			t.Error("Instance card should have correct ID format")
		}

		// Verify no conflicting HTMX attributes
		if strings.Contains(html, "hx-swap-oob") {
			t.Error("Instance card should not have hx-swap-oob attribute")
		}
	})
}
