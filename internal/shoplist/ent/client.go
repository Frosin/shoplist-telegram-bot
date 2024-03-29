// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/migrate"

	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/item"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/shop"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/shopping"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/user"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Item is the client for interacting with the Item builders.
	Item *ItemClient
	// Shop is the client for interacting with the Shop builders.
	Shop *ShopClient
	// Shopping is the client for interacting with the Shopping builders.
	Shopping *ShoppingClient
	// User is the client for interacting with the User builders.
	User *UserClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	cfg := config{log: log.Println, hooks: &hooks{}}
	cfg.options(opts...)
	client := &Client{config: cfg}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Item = NewItemClient(c.config)
	c.Shop = NewShopClient(c.config)
	c.Shopping = NewShoppingClient(c.config)
	c.User = NewUserClient(c.config)
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Item:     NewItemClient(cfg),
		Shop:     NewShopClient(cfg),
		Shopping: NewShoppingClient(cfg),
		User:     NewUserClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Item:     NewItemClient(cfg),
		Shop:     NewShopClient(cfg),
		Shopping: NewShoppingClient(cfg),
		User:     NewUserClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Item.
//		Query().
//		Count(ctx)
//
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Item.Use(hooks...)
	c.Shop.Use(hooks...)
	c.Shopping.Use(hooks...)
	c.User.Use(hooks...)
}

// ItemClient is a client for the Item schema.
type ItemClient struct {
	config
}

// NewItemClient returns a client for the Item from the given config.
func NewItemClient(c config) *ItemClient {
	return &ItemClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `item.Hooks(f(g(h())))`.
func (c *ItemClient) Use(hooks ...Hook) {
	c.hooks.Item = append(c.hooks.Item, hooks...)
}

// Create returns a builder for creating a Item entity.
func (c *ItemClient) Create() *ItemCreate {
	mutation := newItemMutation(c.config, OpCreate)
	return &ItemCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Item entities.
func (c *ItemClient) CreateBulk(builders ...*ItemCreate) *ItemCreateBulk {
	return &ItemCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Item.
func (c *ItemClient) Update() *ItemUpdate {
	mutation := newItemMutation(c.config, OpUpdate)
	return &ItemUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *ItemClient) UpdateOne(i *Item) *ItemUpdateOne {
	mutation := newItemMutation(c.config, OpUpdateOne, withItem(i))
	return &ItemUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *ItemClient) UpdateOneID(id int) *ItemUpdateOne {
	mutation := newItemMutation(c.config, OpUpdateOne, withItemID(id))
	return &ItemUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Item.
func (c *ItemClient) Delete() *ItemDelete {
	mutation := newItemMutation(c.config, OpDelete)
	return &ItemDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *ItemClient) DeleteOne(i *Item) *ItemDeleteOne {
	return c.DeleteOneID(i.ID)
}

// DeleteOne returns a builder for deleting the given entity by its id.
func (c *ItemClient) DeleteOneID(id int) *ItemDeleteOne {
	builder := c.Delete().Where(item.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &ItemDeleteOne{builder}
}

// Query returns a query builder for Item.
func (c *ItemClient) Query() *ItemQuery {
	return &ItemQuery{
		config: c.config,
	}
}

// Get returns a Item entity by its id.
func (c *ItemClient) Get(ctx context.Context, id int) (*Item, error) {
	return c.Query().Where(item.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *ItemClient) GetX(ctx context.Context, id int) *Item {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryShopping queries the shopping edge of a Item.
func (c *ItemClient) QueryShopping(i *Item) *ShoppingQuery {
	query := &ShoppingQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := i.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(item.Table, item.FieldID, id),
			sqlgraph.To(shopping.Table, shopping.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, item.ShoppingTable, item.ShoppingColumn),
		)
		fromV = sqlgraph.Neighbors(i.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *ItemClient) Hooks() []Hook {
	return c.hooks.Item
}

// ShopClient is a client for the Shop schema.
type ShopClient struct {
	config
}

// NewShopClient returns a client for the Shop from the given config.
func NewShopClient(c config) *ShopClient {
	return &ShopClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `shop.Hooks(f(g(h())))`.
func (c *ShopClient) Use(hooks ...Hook) {
	c.hooks.Shop = append(c.hooks.Shop, hooks...)
}

// Create returns a builder for creating a Shop entity.
func (c *ShopClient) Create() *ShopCreate {
	mutation := newShopMutation(c.config, OpCreate)
	return &ShopCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Shop entities.
func (c *ShopClient) CreateBulk(builders ...*ShopCreate) *ShopCreateBulk {
	return &ShopCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Shop.
func (c *ShopClient) Update() *ShopUpdate {
	mutation := newShopMutation(c.config, OpUpdate)
	return &ShopUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *ShopClient) UpdateOne(s *Shop) *ShopUpdateOne {
	mutation := newShopMutation(c.config, OpUpdateOne, withShop(s))
	return &ShopUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *ShopClient) UpdateOneID(id int) *ShopUpdateOne {
	mutation := newShopMutation(c.config, OpUpdateOne, withShopID(id))
	return &ShopUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Shop.
func (c *ShopClient) Delete() *ShopDelete {
	mutation := newShopMutation(c.config, OpDelete)
	return &ShopDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *ShopClient) DeleteOne(s *Shop) *ShopDeleteOne {
	return c.DeleteOneID(s.ID)
}

// DeleteOne returns a builder for deleting the given entity by its id.
func (c *ShopClient) DeleteOneID(id int) *ShopDeleteOne {
	builder := c.Delete().Where(shop.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &ShopDeleteOne{builder}
}

// Query returns a query builder for Shop.
func (c *ShopClient) Query() *ShopQuery {
	return &ShopQuery{
		config: c.config,
	}
}

// Get returns a Shop entity by its id.
func (c *ShopClient) Get(ctx context.Context, id int) (*Shop, error) {
	return c.Query().Where(shop.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *ShopClient) GetX(ctx context.Context, id int) *Shop {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryShopping queries the shopping edge of a Shop.
func (c *ShopClient) QueryShopping(s *Shop) *ShoppingQuery {
	query := &ShoppingQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := s.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(shop.Table, shop.FieldID, id),
			sqlgraph.To(shopping.Table, shopping.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, shop.ShoppingTable, shop.ShoppingColumn),
		)
		fromV = sqlgraph.Neighbors(s.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *ShopClient) Hooks() []Hook {
	return c.hooks.Shop
}

// ShoppingClient is a client for the Shopping schema.
type ShoppingClient struct {
	config
}

// NewShoppingClient returns a client for the Shopping from the given config.
func NewShoppingClient(c config) *ShoppingClient {
	return &ShoppingClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `shopping.Hooks(f(g(h())))`.
func (c *ShoppingClient) Use(hooks ...Hook) {
	c.hooks.Shopping = append(c.hooks.Shopping, hooks...)
}

// Create returns a builder for creating a Shopping entity.
func (c *ShoppingClient) Create() *ShoppingCreate {
	mutation := newShoppingMutation(c.config, OpCreate)
	return &ShoppingCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Shopping entities.
func (c *ShoppingClient) CreateBulk(builders ...*ShoppingCreate) *ShoppingCreateBulk {
	return &ShoppingCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Shopping.
func (c *ShoppingClient) Update() *ShoppingUpdate {
	mutation := newShoppingMutation(c.config, OpUpdate)
	return &ShoppingUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *ShoppingClient) UpdateOne(s *Shopping) *ShoppingUpdateOne {
	mutation := newShoppingMutation(c.config, OpUpdateOne, withShopping(s))
	return &ShoppingUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *ShoppingClient) UpdateOneID(id int) *ShoppingUpdateOne {
	mutation := newShoppingMutation(c.config, OpUpdateOne, withShoppingID(id))
	return &ShoppingUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Shopping.
func (c *ShoppingClient) Delete() *ShoppingDelete {
	mutation := newShoppingMutation(c.config, OpDelete)
	return &ShoppingDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *ShoppingClient) DeleteOne(s *Shopping) *ShoppingDeleteOne {
	return c.DeleteOneID(s.ID)
}

// DeleteOne returns a builder for deleting the given entity by its id.
func (c *ShoppingClient) DeleteOneID(id int) *ShoppingDeleteOne {
	builder := c.Delete().Where(shopping.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &ShoppingDeleteOne{builder}
}

// Query returns a query builder for Shopping.
func (c *ShoppingClient) Query() *ShoppingQuery {
	return &ShoppingQuery{
		config: c.config,
	}
}

// Get returns a Shopping entity by its id.
func (c *ShoppingClient) Get(ctx context.Context, id int) (*Shopping, error) {
	return c.Query().Where(shopping.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *ShoppingClient) GetX(ctx context.Context, id int) *Shopping {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryItem queries the item edge of a Shopping.
func (c *ShoppingClient) QueryItem(s *Shopping) *ItemQuery {
	query := &ItemQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := s.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(shopping.Table, shopping.FieldID, id),
			sqlgraph.To(item.Table, item.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, shopping.ItemTable, shopping.ItemColumn),
		)
		fromV = sqlgraph.Neighbors(s.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryShop queries the shop edge of a Shopping.
func (c *ShoppingClient) QueryShop(s *Shopping) *ShopQuery {
	query := &ShopQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := s.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(shopping.Table, shopping.FieldID, id),
			sqlgraph.To(shop.Table, shop.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, shopping.ShopTable, shopping.ShopColumn),
		)
		fromV = sqlgraph.Neighbors(s.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryUser queries the user edge of a Shopping.
func (c *ShoppingClient) QueryUser(s *Shopping) *UserQuery {
	query := &UserQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := s.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(shopping.Table, shopping.FieldID, id),
			sqlgraph.To(user.Table, user.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, shopping.UserTable, shopping.UserColumn),
		)
		fromV = sqlgraph.Neighbors(s.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *ShoppingClient) Hooks() []Hook {
	return c.hooks.Shopping
}

// UserClient is a client for the User schema.
type UserClient struct {
	config
}

// NewUserClient returns a client for the User from the given config.
func NewUserClient(c config) *UserClient {
	return &UserClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `user.Hooks(f(g(h())))`.
func (c *UserClient) Use(hooks ...Hook) {
	c.hooks.User = append(c.hooks.User, hooks...)
}

// Create returns a builder for creating a User entity.
func (c *UserClient) Create() *UserCreate {
	mutation := newUserMutation(c.config, OpCreate)
	return &UserCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of User entities.
func (c *UserClient) CreateBulk(builders ...*UserCreate) *UserCreateBulk {
	return &UserCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for User.
func (c *UserClient) Update() *UserUpdate {
	mutation := newUserMutation(c.config, OpUpdate)
	return &UserUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *UserClient) UpdateOne(u *User) *UserUpdateOne {
	mutation := newUserMutation(c.config, OpUpdateOne, withUser(u))
	return &UserUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *UserClient) UpdateOneID(id int) *UserUpdateOne {
	mutation := newUserMutation(c.config, OpUpdateOne, withUserID(id))
	return &UserUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for User.
func (c *UserClient) Delete() *UserDelete {
	mutation := newUserMutation(c.config, OpDelete)
	return &UserDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *UserClient) DeleteOne(u *User) *UserDeleteOne {
	return c.DeleteOneID(u.ID)
}

// DeleteOne returns a builder for deleting the given entity by its id.
func (c *UserClient) DeleteOneID(id int) *UserDeleteOne {
	builder := c.Delete().Where(user.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &UserDeleteOne{builder}
}

// Query returns a query builder for User.
func (c *UserClient) Query() *UserQuery {
	return &UserQuery{
		config: c.config,
	}
}

// Get returns a User entity by its id.
func (c *UserClient) Get(ctx context.Context, id int) (*User, error) {
	return c.Query().Where(user.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *UserClient) GetX(ctx context.Context, id int) *User {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryShopping queries the shopping edge of a User.
func (c *UserClient) QueryShopping(u *User) *ShoppingQuery {
	query := &ShoppingQuery{config: c.config}
	query.path = func(ctx context.Context) (fromV *sql.Selector, _ error) {
		id := u.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, id),
			sqlgraph.To(shopping.Table, shopping.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.ShoppingTable, user.ShoppingColumn),
		)
		fromV = sqlgraph.Neighbors(u.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *UserClient) Hooks() []Hook {
	return c.hooks.User
}
