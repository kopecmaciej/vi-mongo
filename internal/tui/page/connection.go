package page

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/util"
)

const (
	ConnectionPageId = "Connection"
)

// Connection is a view for connecting to mongodb using tview package
type Connection struct {
	*core.BaseElement
	*core.Flex

	// form is for creating new connection
	form *core.Form

	// list is a list of all available connections
	list *core.List

	style *config.ConnectionStyle

	// function that is called when connection is set
	onSubmit func()
}

// NewConnection creates a new connection view
func NewConnection() *Connection {
	c := &Connection{
		BaseElement: core.NewBaseElement(),
		Flex:        core.NewFlex(),
		form:        core.NewForm(),
		list:        core.NewList(),
	}

	c.SetIdentifier(ConnectionPageId)

	return c
}

// Init overrides the Init function from the BaseElement struct
func (c *Connection) Init(app *core.App) {
	c.App = app

	c.setLayout()
	c.setStyle()
	c.setKeybindings()

	c.handleEvents()
}

func (c *Connection) handleEvents() {
	go c.HandleEvents(ConnectionPageId, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			c.setStyle()
			go c.App.QueueUpdateDraw(func() {
				c.Render()
			})
		}
	})
}

func (c *Connection) setLayout() {
	c.form.SetTitle(" New connection ")
	c.form.SetBorder(true)

	c.list.SetTitle(" Saved connections ")
	c.list.SetBorder(true)
	c.list.ShowSecondaryText(true)
	c.list.SetWrapText(true)
	c.list.SetBorderPadding(1, 1, 1, 1)
	c.list.SetItemGap(1)

	c.form.AddButton("Save", c.saveButtonFunc)
	c.form.AddButton("Cancel", c.cancelButtonFunc)

}

func (c *Connection) setStyle() {
	c.SetStyle(c.App.GetStyles())
	c.form.SetStyle(c.App.GetStyles())
	c.list.SetStyle(c.App.GetStyles())

	c.style = &c.App.GetStyles().Connection

	c.form.SetFieldTextColor(c.style.FormInputColor.Color())
	c.form.SetFieldBackgroundColor(c.style.FormInputBackgroundColor.Color())
	c.form.SetLabelColor(c.style.FormLabelColor.Color())

	globalBackground := c.App.GetStyles().Global.BackgroundColor.Color()
	mainStyle := tcell.StyleDefault.
		Foreground(c.style.ListTextColor.Color()).
		Background(globalBackground)
	c.list.SetMainTextStyle(mainStyle)

	secondaryStyle := tcell.StyleDefault.
		Foreground(c.style.ListSecondaryTextColor.Color()).
		Background(c.style.ListSecondaryBackgroundColor.Color()).
		Italic(true)
	c.list.SetSecondaryTextStyle(secondaryStyle)
	c.list.SetSelectedStyle(tcell.StyleDefault.
		Foreground(c.style.ListSelectedTextColor.Color()).
		Background(c.style.ListSelectedBackgroundColor.Color()))
}

func (c *Connection) setKeybindings() {
	k := c.App.GetKeys()
	c.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Connection.ConnectionForm.SaveConnection, event.Name()):
			c.saveButtonFunc()
			return nil
		case k.Contains(k.Connection.ConnectionForm.FocusList, event.Name()):
			c.App.SetFocus(c.list)
			return nil
		}

		return event
	})

	c.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Connection.ConnectionList.FocusForm, event.Name()):
			c.App.SetFocus(c.form)
			return nil
		case k.Contains(k.Connection.ConnectionList.DeleteConnection, event.Name()):
			c.deleteCurrConnection()
			return nil
		}
		return event
	})
}

func (c *Connection) Render() {
	c.Clear()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)

	if page, _ := c.App.Pages.GetFrontPage(); page == ConnectionPageId {
		if len(c.App.GetConfig().Connections) > 0 {
			c.renderList()
			defer c.App.SetFocus(c.list)
		} else {
			defer c.App.SetFocus(c.form)
		}
	}

	c.renderForm()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)
}

// renderForm renders the form for creating new connection
func (c *Connection) renderForm() *core.Form {
	c.form.Clear(false)

	c.form.AddInputField("Name", "", 40, nil, nil)

	c.form.AddInputField("Url", "mongodb://", 40, nil, nil)

	c.form.AddTextView("Example", "mongodb://username:password@host:port/db", 40, 1, true, false)
	paste := fmt.Sprintf("Type Url (paste - %s) or fill below", c.App.GetKeys().QueryBar.Paste.String())
	c.form.AddTextView("Info", paste, 40, 1, true, false)
	c.form.AddTextView(" ", "-- ----------------------------------------", 40, 1, true, false)
	c.form.AddInputField("Host", "", 40, nil, nil)
	c.form.AddInputField("Port", "", 10, nil, nil)
	c.form.AddInputField("Username", "", 40, nil, nil)
	c.form.AddPasswordField("Password", "", 40, '*', nil)
	c.form.AddInputField("Database", "", 40, nil, nil)
	c.form.AddInputField("Timeout", "5", 10, nil, nil)
	key := fmt.Sprintf("%s or click", c.App.GetKeys().Connection.ConnectionForm.SaveConnection.String())
	c.form.AddTextView("Save with:", key, 30, 1, true, false)

	c.form.GetFormItemByLabel("Url").(*tview.InputField).SetClipboard(util.GetClipboard())
	c.form.GetFormItemByLabel("Host").(*tview.InputField).SetClipboard(util.GetClipboard())
	c.form.GetFormItemByLabel("Port").(*tview.InputField).SetClipboard(util.GetClipboard())
	c.form.GetFormItemByLabel("Username").(*tview.InputField).SetClipboard(util.GetClipboard())
	c.form.GetFormItemByLabel("Password").(*tview.InputField).SetClipboard(util.GetClipboard())
	c.form.GetFormItemByLabel("Database").(*tview.InputField).SetClipboard(util.GetClipboard())

	c.AddItem(c.form, 60, 0, true)

	return c.form
}

// renderList renders the list of all available connections
func (c *Connection) renderList() {
	c.list.Clear()

	for _, conn := range c.App.GetConfig().Connections {
		uri := "uri: " + conn.GetSafeUri()
		c.list.AddItem(conn.Name, uri, 0, func() {
			c.setConnections()
		})
	}

	c.list.AddItem("Click to add new connection", "or by pressing "+c.App.GetKeys().Connection.ConnectionList.FocusForm.String(), 0, func() {
		c.App.SetFocus(c.form)
	})

	c.AddItem(c.list, 50, 0, true)
}

// setConnections sets connections from config file
func (c *Connection) setConnections() {
	if c.list.GetCurrentItem() == c.list.GetItemCount()-1 {
		return
	}
	connName, _ := c.list.GetItemText(c.list.GetCurrentItem())
	err := c.App.GetConfig().SetCurrentConnection(connName)
	if err != nil {
		modal.ShowError(c.App.Pages, "Failed to set current connection", err)
		return
	}
	c.App.GetConfig().CurrentConnection = connName
	if c.onSubmit != nil {
		c.onSubmit()
	}
}

// removeConnection removes connection from config file
func (c *Connection) deleteCurrConnection() error {
	currItem := c.list.GetCurrentItem()
	currConn, _ := c.list.GetItemText(currItem)
	err := c.App.GetConfig().DeleteConnection(currConn)
	if err != nil {
		return err
	}

	c.Render()
	c.list.SetCurrentItem(currItem)

	return nil
}

// saveButtonFunc is a function for saving new connection
func (c *Connection) saveButtonFunc() {
	name := c.form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
	url := c.form.GetFormItemByLabel("Url").(*tview.InputField).GetText()
	timeout := c.form.GetFormItemByLabel("Timeout").(*tview.InputField).GetText()
	intTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		modal.ShowError(c.App.Pages, "Timeout must be a number", err)
		return
	}
	if url != "mongodb://" && strings.Trim(url, " ") != "" {
		if name == "" {
			name = url
		}
		err := c.App.GetConfig().AddConnectionFromUri(&config.MongoConfig{
			Name:    name,
			Uri:     url,
			Timeout: intTimeout,
		})
		if err != nil {
			modal.ShowError(c.App.Pages, "Failed to save connection", err)
			c.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
			return
		}
	} else {
		host := c.form.GetFormItemByLabel("Host").(*tview.InputField).GetText()
		port := c.form.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		intPort, err := strconv.Atoi(port)
		if err != nil {
			modal.ShowError(c.App.Pages, "Port must be a number", err)
			return
		}
		username := c.form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := c.form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		database := c.form.GetFormItemByLabel("Database").(*tview.InputField).GetText()

		if name == "" {
			name = host + ":" + port
		}
		err = c.App.GetConfig().AddConnection(&config.MongoConfig{
			Name:     name,
			Host:     host,
			Port:     intPort,
			Username: username,
			Password: password,
			Database: database,
			Timeout:  intTimeout,
		})
		if err != nil {
			modal.ShowError(c.App.Pages, "Failed to save connection", err)
			return
		}
	}
	c.Render()
	c.list.SetCurrentItem(c.list.GetItemCount())
}

// cancelButtonFunc is a function for canceling the form
func (c *Connection) cancelButtonFunc() {
	c.renderForm()
}

// SetOnSubmitFunc sets callback function
func (c *Connection) SetOnSubmitFunc(onSubmit func()) {
	c.onSubmit = onSubmit
}
