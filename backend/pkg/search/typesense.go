package search

import (
	"context"
	"fmt"

	"bookmark-sync-service/backend/internal/config"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

// Client wraps the Typesense client with additional functionality
type Client struct {
	client *typesense.Client
	config *config.SearchConfig
}

// NewClient creates a new Typesense client
func NewClient(cfg config.SearchConfig) (*Client, error) {
	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)),
		typesense.WithAPIKey(cfg.APIKey),
		typesense.WithConnectionTimeout(config.TypesenseTimeout),
	)

	return &Client{
		client: client,
		config: &cfg,
	}, nil
}

// HealthCheck checks if Typesense is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.Health(1)
	return err
}

// CreateCollection creates a collection in Typesense
func (c *Client) CreateCollection(ctx context.Context, schema *api.CollectionSchema) error {
	_, err := c.client.Collections().Create(schema)
	return err
}

// DeleteCollection deletes a collection in Typesense
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	_, err := c.client.Collection(name).Delete()
	return err
}

// IndexDocument indexes a document in Typesense
func (c *Client) IndexDocument(ctx context.Context, collection string, document interface{}) error {
	_, err := c.client.Collection(collection).Documents().Create(document)
	return err
}

// UpdateDocument updates a document in Typesense
func (c *Client) UpdateDocument(ctx context.Context, collection string, id string, document interface{}) error {
	_, err := c.client.Collection(collection).Document(id).Update(document)
	return err
}

// DeleteDocument deletes a document in Typesense
func (c *Client) DeleteDocument(ctx context.Context, collection string, id string) error {
	_, err := c.client.Collection(collection).Document(id).Delete()
	return err
}

// Search searches for documents in Typesense
func (c *Client) Search(ctx context.Context, collection string, searchParams *api.SearchCollectionParams) (*api.SearchResult, error) {
	return c.client.Collection(collection).Documents().Search(searchParams)
}

// CreateBookmarkCollection creates the bookmarks collection with Chinese language support
func (c *Client) CreateBookmarkCollection(ctx context.Context) error {
	truePtr := true
	zhPtr := "zh"
	enPtr := "en"
	saveCountPtr := "save_count"

	schema := &api.CollectionSchema{
		Name: "bookmarks",
		Fields: []api.Field{
			{
				Name:  "id",
				Type:  "string",
				Index: &truePtr,
			},
			{
				Name:  "user_id",
				Type:  "string",
				Index: &truePtr,
			},
			{
				Name:   "title",
				Type:   "string",
				Index:  &truePtr,
				Locale: &zhPtr,
			},
			{
				Name:   "description",
				Type:   "string",
				Index:  &truePtr,
				Locale: &zhPtr,
			},
			{
				Name:   "url",
				Type:   "string",
				Index:  &truePtr,
				Locale: &enPtr,
			},
			{
				Name:  "tags",
				Type:  "string[]",
				Index: &truePtr,
				Facet: &truePtr,
			},
			{
				Name:  "created_at",
				Type:  "int64",
				Index: &truePtr,
			},
			{
				Name:  "updated_at",
				Type:  "int64",
				Index: &truePtr,
			},
			{
				Name:  "save_count",
				Type:  "int32",
				Index: &truePtr,
			},
		},
		DefaultSortingField: &saveCountPtr,
	}

	return c.CreateCollection(ctx, schema)
}

// CreateCollectionCollection creates the collections collection with Chinese language support
func (c *Client) CreateCollectionCollection(ctx context.Context) error {
	truePtr := true
	zhPtr := "zh"
	bookmarkCountPtr := "bookmark_count"

	schema := &api.CollectionSchema{
		Name: "collections",
		Fields: []api.Field{
			{
				Name:  "id",
				Type:  "string",
				Index: &truePtr,
			},
			{
				Name:  "user_id",
				Type:  "string",
				Index: &truePtr,
			},
			{
				Name:   "name",
				Type:   "string",
				Index:  &truePtr,
				Locale: &zhPtr,
			},
			{
				Name:   "description",
				Type:   "string",
				Index:  &truePtr,
				Locale: &zhPtr,
			},
			{
				Name:  "visibility",
				Type:  "string",
				Index: &truePtr,
				Facet: &truePtr,
			},
			{
				Name:  "created_at",
				Type:  "int64",
				Index: &truePtr,
			},
			{
				Name:  "updated_at",
				Type:  "int64",
				Index: &truePtr,
			},
			{
				Name:  "bookmark_count",
				Type:  "int32",
				Index: &truePtr,
			},
		},
		DefaultSortingField: &bookmarkCountPtr,
	}

	return c.CreateCollection(ctx, schema)
}

// IndexBookmark indexes a bookmark in Typesense
func (c *Client) IndexBookmark(ctx context.Context, bookmark interface{}) error {
	return c.IndexDocument(ctx, "bookmarks", bookmark)
}

// IndexCollection indexes a collection in Typesense
func (c *Client) IndexCollection(ctx context.Context, collection interface{}) error {
	return c.IndexDocument(ctx, "collections", collection)
}

// SearchBookmarks searches for bookmarks in Typesense
func (c *Client) SearchBookmarks(ctx context.Context, query string, userID string, filters map[string]string, page, limit int) (*api.SearchResult, error) {
	filterBy := fmt.Sprintf("user_id:%s", userID)

	for key, value := range filters {
		filterBy += fmt.Sprintf(" && %s:%s", key, value)
	}

	searchParams := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "title,description,url,tags",
		FilterBy: &filterBy,
		Page:     &page,
		PerPage:  &limit,
	}

	return c.Search(ctx, "bookmarks", searchParams)
}

// SearchCollections searches for collections in Typesense
func (c *Client) SearchCollections(ctx context.Context, query string, userID string, filters map[string]string, page, limit int) (*api.SearchResult, error) {
	filterBy := fmt.Sprintf("user_id:%s", userID)

	for key, value := range filters {
		filterBy += fmt.Sprintf(" && %s:%s", key, value)
	}

	searchParams := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "name,description",
		FilterBy: &filterBy,
		Page:     &page,
		PerPage:  &limit,
	}

	return c.Search(ctx, "collections", searchParams)
}
