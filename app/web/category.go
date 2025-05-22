package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/omegaatt36/bookly/app/web/templates"
)

// category represents a category for web display
type category struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	UserID    int32     `json:"user_id"` // Might not be displayed but useful for logic
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// pageCategories fetches categories via API client and renders category_list.html.
func (s *Server) pageCategories(w http.ResponseWriter, r *http.Request) {
	var categories []category
	err := s.sendRequest(r.Context(), http.MethodGet, "/v1/categories", nil, &categories)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch categories: %w", err))
		return
	}

	s.renderPage(w, r, templates.CategoriesPage(categories, s.getCSRFToken(r)))
}

// pageCreateCategory renders the create_category.html page.
func (s *Server) pageCreateCategory(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, r, templates.CreateCategoryPage(s.getCSRFToken(r)))
}

// createCategory handles form submission from create_category.html,
// calls API to create category, and redirects or updates.
func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	name := r.FormValue("name")
	if name == "" {
		// Handle error: name is required, perhaps re-render form with error
		s.sessionManager.Put(r.Context(), "error", "Category name is required.")
		http.Redirect(w, r, "/categories/create", http.StatusSeeOther)
		return
	}

	payload := map[string]string{"name": name}
	payloadBytes, _ := json.Marshal(payload)

	var createdCategory category
	err := s.sendRequest(r.Context(), http.MethodPost, "/v1/categories", bytes.NewBuffer(payloadBytes), &createdCategory)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to create category: %s", err.Error()))
		// Consider re-rendering the form with the error and previously entered values
		http.Redirect(w, r, "/categories/create", http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Category created successfully.")
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

// pageEditCategory fetches a category and renders edit_category.html.
func (s *Server) pageEditCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // Assuming router supports PathValue for path parameters
	categoryID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid category ID: %w", err))
		return
	}

	var cat category
	err = s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/categories/%d", categoryID), nil, &cat)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch category %d: %w", categoryID, err))
		return
	}

	s.renderPage(w, r, templates.EditCategoryPage(&cat, s.getCSRFToken(r)))
}

// updateCategory handles form submission from edit_category.html, calls API.
func (s *Server) updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	categoryID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid category ID: %w", err))
		return
	}

	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	name := r.FormValue("name")
	if name == "" {
		// Handle error, perhaps re-render form with error and existing data
		s.sessionManager.Put(r.Context(), "error", "Category name is required.")
		// Fetch category again to re-render edit page
		var cat category
		apiErr := s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/categories/%d", categoryID), nil, &cat)
		if apiErr != nil {
			s.handleError(w,r, fmt.Errorf("failed to fetch category for re-rendering edit page: %w", apiErr))
			return
		}
		// This is a simplified error handling. Ideally, you'd pass the error to the template.
		s.renderPage(w, r, templates.EditCategoryPage(&cat, s.getCSRFToken(r)))
		return
	}

	payload := map[string]string{"name": name}
	payloadBytes, _ := json.Marshal(payload)

	// The API uses PUT for update on /v1/categories/{category_id}
	err = s.sendRequest(r.Context(), http.MethodPut, fmt.Sprintf("/v1/categories/%d", categoryID), bytes.NewBuffer(payloadBytes), nil) // Assuming API returns no content or updated category
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to update category: %s", err.Error()))
		// Re-render form with error
		var cat category
        s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/categories/%d", categoryID), nil, &cat) // Fetch fresh data
		s.renderPage(w, r, templates.EditCategoryPage(&cat, s.getCSRFToken(r)))
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Category updated successfully.")
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

// deleteCategory handles deletion, calls API.
func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // Assuming router supports PathValue for path parameters
	categoryID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid category ID: %w", err))
		return
	}

	// CSRF protection check (assuming form submission or specific header for HTMX delete)
	// For simple POST based delete:
	// if r.Method == http.MethodPost {
	//     if err := r.ParseForm(); err != nil {
	//         s.handleError(w, r, fmt.Errorf("failed to parse form: %w", err))
	//         return
	//     }
	//     // Check CSRF token from form if needed
	// }


	err = s.sendRequest(r.Context(), http.MethodDelete, fmt.Sprintf("/v1/categories/%d", categoryID), nil, nil)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to delete category: %s", err.Error()))
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Category deleted successfully.")
	// For HTMX, might return HX-Trigger: categoryListChanged or just 200 OK
	// For standard forms, redirect:
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}
