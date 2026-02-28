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

	// edit mode tracking
	isEditMode      bool
	editingConnName string
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
	c.updateFormTitle()
	c.form.SetBorder(true)

	c.list.SetTitle(" Saved connections ")
	c.list.SetBorder(true)
	c.list.ShowSecondaryText(true)
	c.list.SetWrapText(true)
	c.list.SetBorderPadding(1, 1, 1, 1)
	c.list.SetItemGap(1)

	c.updateFormButtons()
}

func (c *Connection) updateFormTitle() {
	if c.isEditMode {
		c.form.SetTitle(" Edit connection ")
	} else {
		c.form.SetTitle(" Add new connection ")
	}
}

func (c *Connection) updateFormButtons() {
	c.form.ClearButtons()
	var buttonTxt string
	if c.isEditMode {
		buttonTxt = "Update"
	} else {
		buttonTxt = "Save"
	}

	c.form.AddButton(buttonTxt, c.saveButtonFunc)
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
			_, buttonIdx := c.form.GetFocusedItemIndex()

			if buttonIdx >= 0 && buttonIdx < c.form.GetButtonCount() {
				b := c.form.GetButton(buttonIdx)
				if b.GetLabel() == "Cancel" {
					return event
				}
			}

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
		case k.Contains(k.Connection.ConnectionList.MoveUp, event.Name()):
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case k.Contains(k.Connection.ConnectionList.MoveDown, event.Name()):
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case k.Contains(k.Connection.ConnectionList.FocusForm, event.Name()):
			c.App.SetFocus(c.form)
			return nil
		case k.Contains(k.Connection.ConnectionList.DeleteConnection, event.Name()):
			c.deleteCurrConnection()
			return nil
		case k.Contains(k.Connection.ConnectionList.EditConnection, event.Name()):
			c.editCurrConnection()
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
	c.form.Clear(true)

	c.updateFormButtons()

	c.form.AddInputField("Name", "", 40, nil, nil)

	c.form.AddTextArea("Uri", "mongodb://", 40, 3, 0, nil)

	c.form.AddTextView("Example", "mongodb://username:password@host:port/db", 40, 1, true, false)
	paste := fmt.Sprintf("Type Uri (paste - %s) or fill below", c.App.GetKeys().QueryBar.Paste.String())
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

	c.form.GetFormItemByLabel("Uri").(*tview.TextArea).SetClipboard(util.GetClipboard())
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

	editKey := c.App.GetKeys().Connection.ConnectionList.EditConnection.String()
	deleteKey := c.App.GetKeys().Connection.ConnectionList.DeleteConnection.String()
	focusFormKey := c.App.GetKeys().Connection.ConnectionList.FocusForm.String()

	helpText := fmt.Sprintf("Edit (%s) | Delete (%s) | Add new (%s)", editKey, deleteKey, focusFormKey)
	c.list.AddItem("Click to add new connection", helpText, 0, func() {
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

// editCurrConnection loads the selected connection for editing
func (c *Connection) editCurrConnection() {
	currItem := c.list.GetCurrentItem()
	if currItem == c.list.GetItemCount()-1 {
		// This is the "add new connection" item, don't edit it
		return
	}

	connName, _ := c.list.GetItemText(currItem)
	conn, err := c.App.GetConfig().GetConnectionByName(connName)
	if err != nil {
		modal.ShowError(c.App.Pages, "Failed to load connection for editing", err)
		return
	}

	// Enter edit mode
	c.isEditMode = true
	c.editingConnName = connName
	c.updateFormTitle()
	c.updateFormButtons()

	// Populate form with connection data
	c.populateFormWithConnection(conn)

	c.App.SetFocus(c.form)
}

// populateFormWithConnection fills the form with connection data
func (c *Connection) populateFormWithConnection(conn *config.MongoConfig) {
	c.form.GetFormItemByLabel("Name").(*tview.InputField).SetText(conn.Name)

	if conn.Uri != "" && conn.Uri != "mongodb://" {
		c.form.GetFormItemByLabel("Uri").(*tview.TextArea).SetText(conn.Uri, true)
	} else {
		c.form.GetFormItemByLabel("Host").(*tview.InputField).SetText(conn.Host)
		if conn.Port > 0 {
			c.form.GetFormItemByLabel("Port").(*tview.InputField).SetText(fmt.Sprintf("%d", conn.Port))
		}
		c.form.GetFormItemByLabel("Username").(*tview.InputField).SetText(conn.Username)
		c.form.GetFormItemByLabel("Password").(*tview.InputField).SetText(conn.Password)
		c.form.GetFormItemByLabel("Database").(*tview.InputField).SetText(conn.Database)
	}

	if conn.Timeout > 0 {
		c.form.GetFormItemByLabel("Timeout").(*tview.InputField).SetText(fmt.Sprintf("%d", conn.Timeout))
	}
}

// saveButtonFunc handles both saving new connections and updating existing ones
func (c *Connection) saveButtonFunc() {
	name := c.form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
	uri := c.form.GetFormItemByLabel("Uri").(*tview.TextArea).GetText()
	timeout := c.form.GetFormItemByLabel("Timeout").(*tview.InputField).GetText()
	intTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		modal.ShowError(c.App.Pages, "Timeout must be a number", err)
		return
	}

	var saveErr error

	if uri != "mongodb://" && strings.Trim(uri, " ") != "" {
		if name == "" {
			name = uri
		}
		mongoConfig := &config.MongoConfig{
			Name:    name,
			Uri:     uri,
			Timeout: intTimeout,
		}

		if c.isEditMode {
			saveErr = c.App.GetConfig().UpdateConnectionFromUri(c.editingConnName, mongoConfig)
		} else {
			saveErr = c.App.GetConfig().AddConnectionFromUri(mongoConfig)
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
		mongoConfig := &config.MongoConfig{
			Name:     name,
			Host:     host,
			Port:     intPort,
			Username: username,
			Password: password,
			Database: database,
			Timeout:  intTimeout,
		}

		if c.isEditMode {
			saveErr = c.App.GetConfig().UpdateConnection(c.editingConnName, mongoConfig)
		} else {
			saveErr = c.App.GetConfig().AddConnection(mongoConfig)
		}
	}

	if saveErr != nil {
		action := "save"
		if c.isEditMode {
			action = "update"
		}
		modal.ShowError(c.App.Pages, fmt.Sprintf("Failed to %s connection", action), saveErr)
		if !c.isEditMode {
			c.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
		}
		return
	}

	if c.isEditMode {
		c.isEditMode = false
		c.editingConnName = ""
		c.updateFormTitle()
		c.updateFormButtons()
	}

	c.Render()
	c.list.SetCurrentItem(c.list.GetItemCount())
}

// cancelButtonFunc is a function for canceling the form
func (c *Connection) cancelButtonFunc() {
	c.form.Clear(true)
	c.App.SetFocus(c.list)
	c.Render()
}

func (c *Connection) cancelEditFunc() {
	c.isEditMode = false
	c.editingConnName = ""
	c.updateFormTitle()
	c.updateFormButtons()
	c.App.SetFocus(c.list)
}

// SetOnSubmitFunc sets callback function
func (c *Connection) SetOnSubmitFunc(onSubmit func()) {
	c.onSubmit = onSubmit
}
