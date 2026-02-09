package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nhomble/arch-index/internal/schema"
)

func testIndex() *ArchiveIndex {
	raw := &schema.ArchIndex{
		RepoID:   "Hexagonal-Architecture-DDD",
		Patterns: []string{"Hexagonal Architecture"},
		Components: []schema.Component{
			{
				ID:       "customer-service",
				Name:     "Customer Microservice",
				Layer:    "bounded-context",
				CodeRefs: []string{"Customer/**"},
				NestedAnalysis: "components/customer-service.json",
			},
			{
				ID:       "order-service",
				Name:     "Order Microservice",
				Layer:    "bounded-context",
				CodeRefs: []string{"Order/**"},
			},
		},
		Archetypes: map[string][]schema.Archetype{
			"controllers": {
				{
					ID:         "customer-controller",
					File:       "Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/application/rest/controller/CustomerController.java",
					Symbol:     "CustomerController",
					Technology: "spring-mvc",
				},
				{
					ID:         "order-controller",
					File:       "Order/application/src/main/java/com/jmendoza/swa/hexagonal/order/application/rest/controller/OrderController.java",
					Symbol:     "OrderController",
					Technology: "spring-mvc",
				},
			},
			"services": {
				{
					ID:     "create-customer-service",
					File:   "Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/domain/services/CreateCustomerService.java",
					Symbol: "CreateCustomerService",
				},
			},
		},
		Relationships: []schema.Relationship{
			{From: "customer-controller", To: "create-customer-service", Type: "calls", Flow: "create-customer"},
		},
		Flows: []schema.Flow{
			{ID: "create-customer", Name: "Create Customer", Steps: []string{"customer-controller", "create-customer-service"}},
		},
	}
	return NewIndex(raw)
}

func TestHealthEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected ok, got %s", resp["status"])
	}
}

func TestContextEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	// Test with a file inside Customer component
	req := httptest.NewRequest("GET", "/context?file=Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/application/rest/controller/CustomerController.java", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ContextResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Component == nil {
		t.Fatal("expected component, got nil")
	}
	if resp.Component.ID != "customer-service" {
		t.Fatalf("expected customer-service, got %s", resp.Component.ID)
	}
	if resp.Layer != "bounded-context" {
		t.Fatalf("expected bounded-context, got %s", resp.Layer)
	}
	if resp.Archetype == nil {
		t.Fatal("expected archetype, got nil")
	}
	if resp.Archetype.Category != "controllers" {
		t.Fatalf("expected controllers, got %s", resp.Archetype.Category)
	}
	if resp.Archetype.ID != "customer-controller" {
		t.Fatalf("expected customer-controller, got %s", resp.Archetype.ID)
	}
	if !resp.ZoomAvailable {
		t.Fatal("expected zoom_available=true")
	}
}

func TestContextEndpointMissingFile(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/context", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestComponentsEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/components", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string][]schema.Component
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp["components"]) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resp["components"]))
	}
}

func TestArchetypesEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/archetypes/controllers", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Test nonexistent category
	req = httptest.NewRequest("GET", "/archetypes/nonexistent", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRelationshipsEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/relationships?symbol=customer-controller&direction=downstream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string][]schema.Relationship
	json.Unmarshal(w.Body.Bytes(), &resp)
	rels := resp["relationships"]
	if len(rels) != 1 {
		t.Fatalf("expected 1 downstream relationship, got %d", len(rels))
	}
	if rels[0].To != "create-customer-service" {
		t.Fatalf("expected create-customer-service, got %s", rels[0].To)
	}
}

func TestFlowsEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/flows?through=customer-controller", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string][]schema.Flow
	json.Unmarshal(w.Body.Bytes(), &resp)
	flows := resp["flows"]
	if len(flows) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(flows))
	}
	if flows[0].ID != "create-customer" {
		t.Fatalf("expected create-customer, got %s", flows[0].ID)
	}
}

func TestGraphEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/graph", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp GraphPayload
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal graph payload: %v", err)
	}

	if resp.RepoID != "Hexagonal-Architecture-DDD" {
		t.Fatalf("expected repo_id Hexagonal-Architecture-DDD, got %s", resp.RepoID)
	}

	if len(resp.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resp.Components))
	}

	// customer-service should have archetypes nested
	var customerComp *GraphComponent
	for i := range resp.Components {
		if resp.Components[i].ID == "customer-service" {
			customerComp = &resp.Components[i]
			break
		}
	}
	if customerComp == nil {
		t.Fatal("expected customer-service component")
	}
	if len(customerComp.Archetypes) != 2 {
		t.Fatalf("expected 2 archetypes in customer-service, got %d", len(customerComp.Archetypes))
	}

	if len(resp.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(resp.Relationships))
	}

	if len(resp.Flows) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(resp.Flows))
	}

	// component_edges: customer-controller calls create-customer-service,
	// both in customer-service, so no cross-component edge expected.
	if len(resp.ComponentEdges) != 0 {
		t.Fatalf("expected 0 component edges (intra-component), got %d", len(resp.ComponentEdges))
	}
}

func TestUIEndpoint(t *testing.T) {
	idx := testIndex()
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, NewCursorState(idx))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html content type, got %s", ct)
	}

	body := w.Body.String()
	if len(body) < 100 {
		t.Fatal("expected substantial HTML body")
	}
}

func TestCursorPutEndpoint(t *testing.T) {
	idx := testIndex()
	cs := NewCursorState(idx)
	mux := http.NewServeMux()
	SetupRoutes(mux, idx, cs)

	// PUT with a known file
	req := httptest.NewRequest("PUT", "/cursor?file=Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/application/rest/controller/CustomerController.java", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// PUT without file param → 400
	req = httptest.NewRequest("PUT", "/cursor", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCursorSSEBroadcast(t *testing.T) {
	idx := testIndex()
	cs := NewCursorState(idx)

	// Subscribe before setting
	ch := cs.Subscribe()
	defer cs.Unsubscribe(ch)

	cs.Set("Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/application/rest/controller/CustomerController.java")

	msg := <-ch
	var ev CursorEvent
	if err := json.Unmarshal([]byte(msg), &ev); err != nil {
		t.Fatalf("failed to unmarshal cursor event: %v", err)
	}

	if ev.ComponentID != "customer-service" {
		t.Fatalf("expected component_id customer-service, got %s", ev.ComponentID)
	}
	if ev.ArchetypeID != "customer-controller" {
		t.Fatalf("expected archetype_id customer-controller, got %s", ev.ArchetypeID)
	}
}

func TestGlobMatching(t *testing.T) {
	idx := testIndex()

	// File inside Customer/** should match customer-service
	comp := idx.FindComponent("Customer/src/main/java/SomeFile.java")
	if comp == nil {
		t.Fatal("expected component match for Customer/...")
	}
	if comp.ID != "customer-service" {
		t.Fatalf("expected customer-service, got %s", comp.ID)
	}

	// File inside Order/** should match order-service
	comp = idx.FindComponent("Order/domain/SomeFile.java")
	if comp == nil {
		t.Fatal("expected component match for Order/...")
	}
	if comp.ID != "order-service" {
		t.Fatalf("expected order-service, got %s", comp.ID)
	}

	// Unknown file should return nil
	comp = idx.FindComponent("unknown/file.go")
	if comp != nil {
		t.Fatal("expected nil for unknown file")
	}
}
