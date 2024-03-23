<p align="center"><img src="./assets/mongui_header.png"></p>

## Showcasing some of the features of MongUI

### Managing connections (adding, removing)

![](./assets/manage_connections.gif)

### Working with documents (viewing, duplicating, deleting), autocomplete, etc.

![](./assets/working_with_documents.gif)

## Overview

**MongUI** is an intuitive Terminal User Interface (TUI) application, written in
Go, designed to streamline and simplify the management of MongoDB databases.
Emphasizing ease of use without sacrificing functionality, Mongui offers a
user-friendly command-line experience for database administrators and developers
alike.

## Features

- **Intuitive Navigation**: Mongui's simple, intuitive interface makes it easy
  to navigate and manage your MongoDB databases.
- **Managing Documents**: Mongui allows you to view, create, update, duplicate
  and delete documents in your databases with ease.
- **Managing Collections**: Mongui provides a simple way to manage your
  collections, including the ability to create, delete collections.
- **Autocomplete**: Mongui offers an autocomplete feature that suggests
  collection names, database names, and MongoDB commands as you type.
- **Query History**: Mongui keeps track of your query history, allowing you to
  easily access and reuse previous queries.

## Installation

### Using curl

```bash
curl https://api.github.com/repos/kopecmaciej/mongui/releases/latest | jq -r '.assets[0].browser_download_url' | xargs curl -L -o mongui
# or if no jq installed
# curl https://api.github.com/repos/kopecmaciej/mongui/releases/latest | grep browser_download_url | cut -d '"' -f4 | xargs curl -L -o mongui
chmod +x mongui
sudo mv mongui /usr/bin
```

### Using wget

```bash
wget -O - https://api.github.com/repos/kopecmaciej/mongui/releases/latest | jq -r '.assets[0].browser_download_url' | xargs wget
chmod +x mongui
sudo mv mongui /usr/bin
```

### Using Go

```bash
git clone git@github.com:kopecmaciej/mongui.git
cd mongui
make
```

## Usage

After installing Mongui, you can run it by typing `mongui` in your terminal.

```bash
mongui
```

In any moment you can press `Ctrl + H` to see help page with all available
shortcuts. Resizing terminal while running Mongui should work fine, but if you
encounter any issues, please let me know.

All configuration files should be stored in `~/.config/mongui` directory, but it
depends on the system settings as env `XDG_CONFIG_HOME` can be set to different
directories (more information here:
[XDG Base Directory](https://github.com/adrg/xdg?tab=readme-ov-file#xdg-base-directory))

## List of features to be implemented

- [x] Query History
- [x] Switching between multiple Connections
- [x] Help page
- [ ] Improve Content by adding other possibilities of viewing
- [ ] Multiple styles
- [ ] Managing Indexes
- [ ] Aggregation Pipeline
- [ ] Exporting/Importing Documents
- [x] Move autocomplete keys to json file, so that it can be easily modified

## Issues

- [x] Searching collection on databases not expanding tree
- [ ] No clearing history
- [ ] Header not updated after changing database/collection
- [ ] No view with shortcuts
- [ ] Content not updated after editing from picker
- [ ] Performance issue while loading large bson files

```

```
