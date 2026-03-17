1. We have duplicated a lot of keys, example are all `input bars` which works mostly the same, eg:

```
	QueryBar struct {
		ShowHistory Key `yaml:"showHistory"`
		ClearInput  Key `yaml:"clearInput"`
		Paste       Key `yaml:"paste"`
		NextMarker  Key `yaml:"nextMarker"`
		PrevMarker  Key `yaml:"prevMarker"`
	}

	SortBar struct {
		ClearInput Key `yaml:"clearInput"`
		Paste      Key `yaml:"paste"`
	}
```

We should probably have a way to inherit default keybidings like ClearInput or Paste from them. Or maybe it should be keep this way, what you think?

2. We need to add MovementKeys to the keybidings and then apply them when needed. They will be keys like: `left` or `h`, `right` or `l`, `Ctrl+h` or `Tab` to change focus etc. Defined in one place, inherit everywhere. Probably some `tview` changes will be needed as well
