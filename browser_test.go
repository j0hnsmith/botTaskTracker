package main

import (
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// TestBrowserUI tests the UI using rod browser automation
func TestBrowserUI(t *testing.T) {
	// Launch browser
	l := launcher.New().
		Headless(true).
		Devtools(false)
	defer l.Cleanup()

	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	// Navigate to the app (assumes server is running on localhost:7002)
	page := browser.MustPage("http://localhost:7002")
	defer page.MustClose()

	// Wait for page to load
	page.MustWaitLoad()

	t.Run("Board loads with all columns", func(t *testing.T) {
		// Check for column headers
		if !page.MustHas("text=Backlog") {
			t.Error("Backlog column not found")
		}
		if !page.MustHas("text=In Progress") {
			t.Error("In Progress column not found")
		}
		if !page.MustHas("text=Review") {
			t.Error("Review column not found")
		}
		if !page.MustHas("text=Done") {
			t.Error("Done column not found")
		}
	})

	t.Run("Stats are visible in header", func(t *testing.T) {
		// Check for stats component
		if !page.MustHas(".stats") {
			t.Error("Stats component not found in header")
		}
	})

	t.Run("Activity stream is visible", func(t *testing.T) {
		// Check for activity stream section
		if !page.MustHas("text=Activity Stream") {
			t.Error("Activity stream not found")
		}
		
		// Timeline should exist
		if !page.MustHas(".timeline") {
			t.Error("Timeline component not found")
		}
	})

	t.Run("Task cards use daisyUI components", func(t *testing.T) {
		// Check for daisyUI card components
		if !page.MustHas(".card") {
			t.Error("Card component not found")
		}
		
		// Check for badges (tags)
		if !page.MustHas(".badge") {
			t.Error("Badge component not found")
		}
		
		// Check for avatars
		if !page.MustHas(".avatar") {
			t.Error("Avatar component not found")
		}
	})

	t.Run("Progress bars on in-progress tasks", func(t *testing.T) {
		// Check for progress components
		if page.MustHas("text=In Progress") {
			// Navigate to in-progress column area and look for progress bars
			if !page.MustHas(".progress") {
				t.Error("Progress bar component not found in In Progress column")
			}
		}
	})

	t.Run("Dropdown menus work", func(t *testing.T) {
		// Find a task card menu button
		menuBtn := page.MustElement(".dropdown .btn-square")
		if menuBtn == nil {
			t.Error("Dropdown menu button not found")
		}
		
		// Click the menu
		menuBtn.MustClick()
		time.Sleep(200 * time.Millisecond)
		
		// Check if menu items appear
		if !page.MustHas("text=Edit tags") {
			t.Error("Edit tags menu item not found")
		}
		if !page.MustHas("text=Change assignee") {
			t.Error("Change assignee menu item not found")
		}
		if !page.MustHas("text=Delete") {
			t.Error("Delete menu item not found")
		}
	})

	t.Run("Filter dropdown works", func(t *testing.T) {
		// Find filter button
		filterBtn := page.MustElement("text=Filter")
		if filterBtn == nil {
			t.Error("Filter button not found")
		}
		
		// Click filter
		filterBtn.MustClick()
		time.Sleep(200 * time.Millisecond)
		
		// Check for filter options
		if !page.MustHas("text=All assignees") {
			t.Error("Filter dropdown not working")
		}
	})

	t.Run("Breadcrumbs navigation exists", func(t *testing.T) {
		// Check for breadcrumbs
		if !page.MustHas(".breadcrumbs") {
			t.Error("Breadcrumbs component not found")
		}
	})

	t.Run("Swimlane layout is correct", func(t *testing.T) {
		// Check that swimlanes are side-by-side
		swimlanes := page.MustElements(".swimlane")
		if len(swimlanes) != 4 {
			t.Errorf("Expected 4 swimlanes, got %d", len(swimlanes))
		}
	})
}

// TestVisualRegression takes screenshots for visual comparison
func TestVisualRegression(t *testing.T) {
	l := launcher.New().
		Headless(true).
		Devtools(false)
	defer l.Cleanup()

	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("http://localhost:7002")
	defer page.MustClose()
	page.MustWaitLoad()

	// Take full page screenshot
	screenshot, err := page.Screenshot(true, nil)
	if err != nil {
		t.Fatalf("Failed to take screenshot: %v", err)
	}

	// Save screenshot
	err = page.MustScreenshot("test-screenshots/board-full.png")
	if err != nil {
		t.Logf("Screenshot saved (%d bytes)", len(screenshot))
	}

	// Take screenshots of individual columns
	columns := []string{"backlog", "in_progress", "review", "done"}
	for _, col := range columns {
		elem := page.MustElement("#column-" + col)
		if elem != nil {
			elem.MustScreenshot("test-screenshots/column-" + col + ".png")
		}
	}
}
