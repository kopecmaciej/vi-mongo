package component

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	pathToConfig = "config.json"
)

// Connector is a view for connecting to mongodb using tview package
type Connector struct {
	*Component
	*tview.Flex

	// form is for creating new connection
	form *tview.Form

	// list is a list of all available connections
	list *tview.List

	// callback is a function that is called after connection is selected
	callback func()
}

// NewConnector creates a new connection view
func NewConnector() *Connector {
	c := &Connector{
		Component: NewComponent("Connector"),
		Flex:      tview.NewFlex(),
		form:      tview.NewForm(),
		list:      tview.NewList(),
	}

	return c
}

// Init initializes the Component
func (c *Connector) Init(app *App) error {
	c.app = app

	c.setStyle()
	c.setKeybindings()

	c.render(0)

	return nil
}

func (c *Connector) setStyle() {
	style := c.app.Styles.Connector

	c.SetBackgroundColor(style.BackgroundColor.Color())
	c.form.SetBackgroundColor(style.BackgroundColor.Color())
	c.list.SetBackgroundColor(style.BackgroundColor.Color())

	c.form.SetBorder(true)
	c.list.SetBorder(true)
	c.form.SetTitle(" New connection ")
	c.list.SetTitle(" Recent connections ")

	c.list.ShowSecondaryText(true)
	c.list.SetWrapText(true)

	c.form.SetFieldTextColor(style.FormInputColor.Color())
	c.form.SetFieldBackgroundColor(style.FormInputBackgroundColor.Color())

	mainStyle := tcell.StyleDefault.
		Foreground(style.ListTextColor.Color()).
		Background(style.BackgroundColor.Color())
	c.list.SetMainTextStyle(mainStyle)

	secondaryStyle := tcell.StyleDefault.
		Foreground(style.ListSecondaryTextColor.Color()).
		Background(style.BackgroundColor.Color()).
		Italic(true)
	c.list.SetSecondaryTextStyle(secondaryStyle)
}

func (c *Connector) setKeybindings() {
	c.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			if c.moveFormFocusDown() {
				return nil
			}
		case tcell.KeyUp:
			if c.moveFormFocusUp() {
				return nil
			}
		case tcell.KeyCtrlS:
			c.saveButtonFunc()
			return nil
		case tcell.KeyEsc:
			c.app.SetFocus(c.list)
			return nil
		}
		return event
	})
	c.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			c.app.SetFocus(c.list)
			return nil
		case tcell.KeyCtrlA:
			c.app.SetFocus(c.form)
			return nil
		case tcell.KeyCtrlD:
			c.deleteCurrConnection()
			return nil
		case tcell.KeyEnter:
			connName, _ := c.list.GetItemText(c.list.GetCurrentItem())
			err := c.app.Config.SetCurrentConnection(connName)
			if err != nil {
				log.Error().Err(err).Msg("failed to set current connection")
			}
			c.app.Config.CurrentConnection = connName
			c.app.Root.RemovePage("Connector")
			if c.callback != nil {
				c.callback()
			}
			return nil
		}
		return event
	})

}

// render renders the Component
func (c *Connector) render(currItem int) {
	c.Clear()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)

	if len(c.app.Config.Connections) > 0 {
		c.renderList(currItem)
	}
	c.renderForm()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)

	c.app.SetFocus(c.list)
}

// renderForm renders the form for creating new connection
func (c *Connector) renderForm() *tview.Form {
	c.form.Clear(true)

	c.form.AddInputField("Name", "", 40, nil, nil)
	c.form.AddInputField("Url", "", 40, nil, nil)
	c.form.AddTextView("Info", "You can either provide a connection string or fill the form below", 40, 2, true, false)
	c.form.AddInputField("Host", "", 40, nil, nil)
	c.form.AddInputField("Port", "", 10, nil, nil)
	c.form.AddInputField("Username", "", 40, nil, nil)
	c.form.AddPasswordField("Password", "", 40, '*', nil)
	c.form.AddInputField("Database", "", 40, nil, nil)

	c.form.AddButton("Save", c.saveButtonFunc)
	c.form.AddButton("Cancel", c.cancelButtonFunc)

	c.AddItem(c.form, 60, 0, true)

	return c.form
}

// renderList renders the list of all available connections
func (c *Connector) renderList(currItem int) {
	c.list.Clear()
	// let's add little padding to the list
	c.list.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		return x + 3, y + 2, width, height
	})

	for _, conn := range c.app.Config.Connections {
		uri := "uri: " + conn.GetUri()
		c.list.AddItem(conn.Name, uri, 0, nil)
	}

	button := tview.NewButton("Add new connection")
	button.SetSelectedFunc(func() {
		c.app.SetFocus(c.form)
	})

	c.list.SetCurrentItem(currItem)

	c.AddItem(c.list, 50, 0, true)
}

// saveConnection saves new connection to config file
func (c *Connector) saveConnection(mongoCfg *config.MongoConfig) error {
	err := c.app.Config.AddConnection(mongoCfg)
	if err != nil {
		return err
	}

	return nil
}

// removeConnection removes connection from config file
func (c *Connector) deleteCurrConnection() error {
	currConn, _ := c.list.GetItemText(c.list.GetCurrentItem())
	defer c.render(c.list.GetCurrentItem())
	err := c.app.Config.DeleteConnection(currConn)
	if err != nil {
		return err
	}

	return nil
}

// saveButtonFunc is a function for saving new connection
func (c *Connector) saveButtonFunc() {
	name := c.form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
	url := c.form.GetFormItemByLabel("Url").(*tview.InputField).GetText()
	if url != "" {
		err := c.saveConnection(&config.MongoConfig{
			Name: name,
			Uri:  url,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to save connection")
			ShowErrorModal(c.app.Root, "Failed to save connection")
		}
	} else {
		host := c.form.GetFormItemByLabel("Host").(*tview.InputField).GetText()
		port := c.form.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		intPort, err := strconv.Atoi(port)
		if err != nil {
			log.Error().Err(err).Msg("Port must be a number")
			ShowErrorModal(c.app.Root, "Port must be a number")
		}
		username := c.form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := c.form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		database := c.form.GetFormItemByLabel("Database").(*tview.InputField).GetText()

		err = c.saveConnection(&config.MongoConfig{
			Name:     name,
			Host:     host,
			Port:     intPort,
			Username: username,
			Password: password,
			Database: database,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to save connection")
			ShowErrorModal(c.app.Root, "Failed to save connection")
		}
	}
	c.render(c.list.GetItemCount())
}

// cancelButtonFunc is a function for canceling the form
func (c *Connector) cancelButtonFunc() {
	c.form.Clear(true)
	c.render(0)
}

// SetCallback sets callback function
func (c *Connector) SetCallback(callback func()) {
	c.callback = callback
}

// moveFormFocusUp moves the focus in the form up
func (c *Connector) moveFormFocusUp() bool {
	if c.form.HasFocus() {
		index, _ := c.form.GetFocusedItemIndex()
		// skip the textview item
		if index == 3 {
			index -= 2
		} else {
			index--
		}
		c.form.SetFocus(index)
	}
	return c.form.HasFocus()
}

// moveFormFocusDown moves the focus down in the form
func (c *Connector) moveFormFocusDown() bool {
	if c.form.HasFocus() {
		index, _ := c.form.GetFocusedItemIndex()
		// skip the textview item
		if index == 1 {
			index += 2
		} else {
			index++
		}
		c.form.SetFocus(index)
	}
	return c.form.HasFocus()
}
