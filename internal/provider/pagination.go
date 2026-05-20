// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/url"
)

type kindeRequester interface {
	NewRequest(ctx context.Context, method, path string, query url.Values, payload any) (*http.Request, error)
	DoRequest(req *http.Request, result any) error
}

type kindePage[T any] interface {
	getData() []T
	getNextToken() string
}

func getAllPages[T any, P kindePage[T]](ctx context.Context, client kindeRequester, endpoint string, query url.Values) ([]T, error) {
	if query == nil {
		query = url.Values{}
	}
	if query.Get("page_size") == "" {
		query.Set("page_size", "100")
	}

	var results []T
	for {
		req, err := client.NewRequest(ctx, http.MethodGet, endpoint, query, nil)
		if err != nil {
			return nil, err
		}

		var response P
		if err := client.DoRequest(req, &response); err != nil {
			return nil, err
		}

		results = append(results, response.getData()...)
		nextToken := response.getNextToken()
		if nextToken == "" {
			break
		}
		query.Set("next_token", nextToken)
	}

	return results, nil
}
