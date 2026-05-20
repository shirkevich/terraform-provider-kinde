// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nxt-fwd/kinde-go/api/roles"
	"github.com/nxt-fwd/terraform-provider-kinde/internal/serde"
)

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

//nolint:unused
func expandRoleResourceModel(data RoleResourceModel) roles.Role {
	return roles.Role{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Key:         data.Key.ValueString(),
		Description: data.Description.ValueString(),
	}
}

func expandRoleCreateParams(plan RoleResourceModel) roles.CreateParams {
	return roles.CreateParams{
		Name:        plan.Name.ValueString(),
		Key:         plan.Key.ValueString(),
		Description: plan.Description.ValueString(),
	}
}

func expandRoleUpdateParams(plan RoleResourceModel) roles.UpdateParams {
	return roles.UpdateParams{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
}

func flattenRolePermissions(ctx context.Context, permissions []string, nullWhenEmpty bool) (types.Set, error) {
	if len(permissions) == 0 && nullWhenEmpty {
		return types.SetNull(types.StringType), nil
	}
	if permissions == nil {
		permissions = []string{}
	}

	permissionsSet, diags := types.SetValueFrom(ctx, types.StringType, permissions)
	if diags.HasError() {
		return types.Set{}, fmt.Errorf("failed to flatten permissions: %v", diags)
	}

	return permissionsSet, nil
}

func flattenRoleResource(ctx context.Context, role *roles.Role, permissions []string, nullWhenEmpty bool) (RoleResourceModel, error) {
	permissionsSet, err := flattenRolePermissions(ctx, permissions, nullWhenEmpty)
	if err != nil {
		return RoleResourceModel{}, err
	}

	return RoleResourceModel{
		ID:          types.StringValue(role.ID),
		Name:        types.StringValue(role.Name),
		Key:         types.StringValue(role.Key),
		Description: types.StringValue(role.Description),
		Permissions: permissionsSet,
	}, nil
}

type RoleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

//nolint:unused
func expandRoleDataSourceModel(model RoleDataSourceModel) *roles.Role {
	return &roles.Role{
		ID:          model.ID.ValueString(),
		Name:        model.Name.ValueString(),
		Key:         model.Key.ValueString(),
		Description: model.Description.ValueString(),
	}
}

//nolint:unused
func flattenRoleDataSource(ctx context.Context, resource *roles.Role, permissions []string) (RoleDataSourceModel, error) {
	permissionsList, diags := serde.FlattenStringList(ctx, permissions)
	if diags.HasError() {
		return RoleDataSourceModel{}, fmt.Errorf("failed to flatten permissions: %v", diags)
	}

	return RoleDataSourceModel{
		ID:          types.StringValue(resource.ID),
		Name:        types.StringValue(resource.Name),
		Key:         types.StringValue(resource.Key),
		Description: types.StringValue(resource.Description),
		Permissions: permissionsList,
	}, nil
}
