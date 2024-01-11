package component

import (
	"context"

	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
)

const (
	ConnectorComponent manager.Component = "Connector"

	pathToConfig = "config.json"
)

// Connector is a view for connecting to mongodb using tview package
type Connector struct {
	*Component
	*tview.Flex

	form *tview.Form

	// list of all available connections
	list *tview.List

	// list of all available connections
	connections []*config.MongoConfig
}

// NewConnector creates a new connection view
func NewConnector() *Connector {
	c := &Connector{
		Flex: tview.NewFlex().SetDirection(tview.FlexRow),
		form: tview.NewForm(),
		list: tview.NewList(),
	}

	return c
}

// Init initializes the Component
func (c *Connector) Init(ctx context.Context) error {
	app, err := GetApp(ctx)
	if err != nil {
		return err
	}
	c.app = app

	c.setStyle()

	return nil
}

func (c *Connector) setStyle() {
	c.SetBorder(true)
	c.SetTitle("Connect to MongoDB")
	c.SetTitleAlign(tview.AlignLeft)
	c.SetRect(0, 0, 50, 10)
}

// Render renders the Component
func (c *Connector) Render() {
	c.Clear()

	c.list.Clear()

	for _, conn := range c.connections {
		c.list.AddItem(conn.Name, "", 0, nil)
	}
}

func (c *Connector) DoneFuncHandler() {
}

// ConnectionForm returns a form for creating new connection
func (c *Connector) ConnectionForm() *tview.Form {
	c.form.Clear(true)

	c.form.AddInputField("Name", "", 20, nil, nil)
	c.form.AddInputField("Host", "", 20, nil, nil)
	c.form.AddInputField("Port", "", 20, nil, nil)
	c.form.AddInputField("Username", "", 20, nil, nil)
	c.form.AddPasswordField("Password", "", 20, '*', nil)
	c.form.AddInputField("Database", "", 20, nil, nil)

	c.form.AddButton("Save", func() {
		// c.SaveConnection()
	})
	c.form.AddButton("Cancel", func() {
	})

	return c.form
}

// AddConnection adds a new connection to the list
func (c *Connector) AddConnection(conn *config.MongoConfig) {
	c.connections = append(c.connections, conn)
}

// SaveConnection saves the current connection to the config file
// func (c *Connector) SaveConnection() error {
// 	return config.SaveMongoConfig(c.config)
// }
