package model

import (
	"github.com/kopecmaciej/mongui/internal/manager"
	"github.com/kopecmaciej/mongui/internal/mongo"
	"github.com/kopecmaciej/tview"
)

// AppInterface defines the methods that the App struct should implement
type AppInterface interface {
	GetDao() *mongo.Dao
	GetManager() *manager.ViewManager
	SetFocus(p tview.Primitive)
}
