package handlers

import (
	"net/http"

	"github.com/sidryenireddy/prism/api/internal/mockdata"
)

type MockDataHandler struct{}

func NewMockDataHandler() *MockDataHandler {
	return &MockDataHandler{}
}

func (h *MockDataHandler) ObjectTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, mockdata.ObjectTypes)
}

func (h *MockDataHandler) Objects(w http.ResponseWriter, r *http.Request) {
	otID := r.URL.Query().Get("objectTypeId")
	objects := mockdata.GetObjectsByType(otID)
	if objects == nil {
		objects = []mockdata.MockObject{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"objects":    objects,
		"totalCount": len(objects),
	})
}
