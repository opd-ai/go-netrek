package config

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

func TestGalaxyTemplateSystem(t *testing.T) {
	// Test 1: Verify we can get a galaxy template
	template := GetGalaxyTemplate("classic_netrek")
	if template == nil {
		t.Fatal("Expected to get classic_netrek template, got nil")
	}

	if template.Name != "Classic Netrek" {
		t.Errorf("Expected template name 'Classic Netrek', got '%s'", template.Name)
	}

	if len(template.Planets) != 9 {
		t.Errorf("Expected classic_netrek template to have 9 planets, got %d", len(template.Planets))
	}

	if len(template.Teams) != 4 {
		t.Errorf("Expected classic_netrek template to have 4 teams, got %d", len(template.Teams))
	}

	// Test 2: Verify we can list available templates
	templates := ListGalaxyTemplates()
	if len(templates) == 0 {
		t.Error("Expected to get list of galaxy templates")
	}

	expectedTemplates := []string{"classic_netrek", "small_galaxy", "balanced_4team"}
	for _, expected := range expectedTemplates {
		if _, ok := templates[expected]; !ok {
			t.Errorf("Expected template '%s' to be available", expected)
		}
	}

	// Test 3: Verify we can apply a template to config
	cfg := DefaultConfig()
	err := ApplyGalaxyTemplate(cfg, "small_galaxy")
	if err != nil {
		t.Fatalf("Failed to apply galaxy template: %v", err)
	}

	// Verify template was applied
	if cfg.WorldSize != 6000 {
		t.Errorf("Expected world size 6000 from small_galaxy template, got %f", cfg.WorldSize)
	}

	if len(cfg.Teams) != 2 {
		t.Errorf("Expected 2 teams from small_galaxy template, got %d", len(cfg.Teams))
	}

	if len(cfg.Planets) != 4 {
		t.Errorf("Expected 4 planets from small_galaxy template, got %d", len(cfg.Planets))
	}

	// Test 4: Verify unknown template returns error
	err = ApplyGalaxyTemplate(cfg, "unknown_template")
	if err == nil {
		t.Error("Expected error for unknown template")
	}

	// Test 5: Test LoadConfigWithTemplate function
	cfg2, err := LoadConfigWithTemplate("nonexistent.json", "small_galaxy")
	if err != nil {
		t.Fatalf("LoadConfigWithTemplate should fall back to default config, got error: %v", err)
	}

	if cfg2.WorldSize != 6000 {
		t.Errorf("Expected world size 6000 after template application, got %f", cfg2.WorldSize)
	}
}

func TestGalaxyTemplateValidation(t *testing.T) {
	// Test that all built-in templates are valid
	for name, template := range galaxyTemplates {
		t.Run(name, func(t *testing.T) {
			if template.Name == "" {
				t.Error("Template name should not be empty")
			}

			if template.Description == "" {
				t.Error("Template description should not be empty")
			}

			if template.WorldSize <= 0 {
				t.Error("Template world size should be positive")
			}

			if len(template.Teams) == 0 {
				t.Error("Template should have at least one team")
			}

			if len(template.Planets) == 0 {
				t.Error("Template should have at least one planet")
			}

			// Validate home worlds
			homeWorlds := 0
			for _, planet := range template.Planets {
				if planet.HomeWorld {
					homeWorlds++
					if planet.Type != entity.Homeworld {
						t.Errorf("Home world planet '%s' should have Homeworld type", planet.Name)
					}
					if planet.TeamID < 0 || planet.TeamID >= len(template.Teams) {
						t.Errorf("Home world planet '%s' has invalid team ID %d", planet.Name, planet.TeamID)
					}
				}
			}

			if homeWorlds == 0 {
				t.Error("Template should have at least one home world")
			}
		})
	}
}
