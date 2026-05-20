package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/roles"
)

func TestBuildRolePermissionOperations(t *testing.T) {
	got := buildRolePermissionOperations(
		[]string{"perm_c", "perm_a", "perm_b"},
		[]string{"perm_b", "perm_d"},
	)

	want := []roles.UpdatePermissionItem{
		{ID: "perm_a", Operation: "delete"},
		{ID: "perm_c", Operation: "delete"},
		{ID: "perm_d"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected operations\nwant: %#v\n got: %#v", want, got)
	}
}

func TestRolePermissionsMatchIgnoresOrdering(t *testing.T) {
	if !rolePermissionsMatch([]string{"perm_b", "perm_a"}, []string{"perm_a", "perm_b"}) {
		t.Fatal("expected matching permission sets with different ordering")
	}

	if rolePermissionsMatch([]string{"perm_a"}, []string{"perm_a", "perm_b"}) {
		t.Fatal("expected different permission sets not to match")
	}
}

func TestFlattenRolePermissionsPreservesNullAndEmpty(t *testing.T) {
	ctx := context.Background()

	nullSet, err := flattenRolePermissions(ctx, nil, true)
	if err != nil {
		t.Fatalf("flatten null permissions: %s", err)
	}
	if !nullSet.IsNull() {
		t.Fatalf("expected null permissions, got %#v", nullSet)
	}

	emptySet, err := flattenRolePermissions(ctx, nil, false)
	if err != nil {
		t.Fatalf("flatten empty permissions: %s", err)
	}
	if emptySet.IsNull() {
		t.Fatal("expected configured empty permissions to remain an empty set, got null")
	}
	if len(emptySet.Elements()) != 0 {
		t.Fatalf("expected empty set, got %d elements", len(emptySet.Elements()))
	}
}

func TestGetRoleUsesAllPermissionPages(t *testing.T) {
	ctx := context.Background()
	var permissionRequests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"expires_in":   3600,
				"token_type":   "bearer",
			})
		case "/api/v1/roles/role_1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"role": map[string]any{
					"id":          "role_1",
					"name":        "Role",
					"key":         "role",
					"description": "Role description",
				},
			})
		case "/api/v1/roles/role_1/permissions":
			permissionRequests = append(permissionRequests, r.URL.RawQuery)
			if r.URL.Query().Get("next_token") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"next_token": "page-2",
					"permissions": []map[string]string{
						{"id": "perm_01"},
						{"id": "perm_02"},
						{"id": "perm_03"},
						{"id": "perm_04"},
						{"id": "perm_05"},
						{"id": "perm_06"},
						{"id": "perm_07"},
						{"id": "perm_08"},
						{"id": "perm_09"},
						{"id": "perm_10"},
					},
				})
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"permissions": []map[string]string{
					{"id": "perm_11"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := kinde.New(ctx, kinde.NewClientOptions().
		WithDomain(server.URL).
		WithAudience("audience").
		WithClientID("client-id").
		WithClientSecret("client-secret"))

	resource := RoleResource{client: client.Roles}
	role, err := resource.getRole(ctx, "role_1")
	if err != nil {
		t.Fatalf("get role: %s", err)
	}

	if got := len(role.Permissions); got != 11 {
		t.Fatalf("expected 11 permissions across pages, got %d: %#v", got, role.Permissions)
	}

	if len(permissionRequests) != 2 {
		t.Fatalf("expected 2 permission requests, got %d: %#v", len(permissionRequests), permissionRequests)
	}

	wantQueries := []url.Values{
		{"page_size": []string{"10"}},
		{"page_size": []string{"10"}, "next_token": []string{"page-2"}},
	}
	for i, wantQuery := range wantQueries {
		gotQuery, err := url.ParseQuery(permissionRequests[i])
		if err != nil {
			t.Fatalf("parse permission request query %q: %s", permissionRequests[i], err)
		}
		if !reflect.DeepEqual(gotQuery, wantQuery) {
			t.Fatalf("unexpected permission request query %d\nwant: %#v\n got: %#v", i, wantQuery, gotQuery)
		}
	}
}

func TestFlattenRoleResourceUsesEmptySetForConfiguredEmpty(t *testing.T) {
	state, err := flattenRoleResource(context.Background(), &roles.Role{
		ID:          "role_1",
		Name:        "Role",
		Key:         "role",
		Description: "Role description",
	}, nil, false)
	if err != nil {
		t.Fatalf("flatten role: %s", err)
	}

	if state.Permissions.IsNull() {
		t.Fatal("expected empty configured permissions to stay an empty set")
	}

	emptySet, diags := types.SetValueFrom(context.Background(), types.StringType, []string{})
	if diags.HasError() {
		t.Fatalf("build empty set: %v", diags)
	}
	if !state.Permissions.Equal(emptySet) {
		t.Fatalf("expected empty set, got %#v", state.Permissions)
	}
}
