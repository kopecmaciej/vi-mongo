<p align="center"><img src="./assets/vimongo_header.png"></p>

## Showcasing some of the features of Vi Mongo

### Managing connections (adding, removing)

![](./assets/manage_connections.gif)

### Working with documents (viewing, duplicating, deleting), autocomplete, etc.

![](./assets/working_with_documents.gif)

## Overview

**Vi Mongo** is an intuitive Terminal User Interface (TUI) application, written
in Go, designed to streamline and simplify the management of MongoDB databases.
Emphasizing ease of use without sacrificing functionality, Vi Mongo offers a
user-friendly command-line experience for database administrators and developers
alike.

## Features

- **Intuitive Navigation**: Vi Mongo's simple, intuitive interface makes it easy
  to navigate and manage your MongoDB databases.
- **Managing Documents**: Vi Mongo allows you to view, create, update, duplicate
  and delete documents in your databases with ease.
- **Managing Collections**: Vi Mongo provides a simple way to manage your
  collections, including the ability to create, delete collections.
- **Autocomplete**: Vi Mongo offers an autocomplete feature that suggests
  collection names, database names, and MongoDB commands as you type.
- **Query History**: Vi Mongo keeps track of your query history, allowing you to
  easily access and reuse previous queries.

## Installation

### Using curl

```bash
curl https://api.github.com/repos/kopecmaciej/vimongo/releases/latest | jq -r '.assets[0].browser_download_url' | xargs curl -L -o vimongo
# or if no jq installed
# curl https://api.github.com/repos/kopecmaciej/vimongo/releases/latest | grep browser_download_url | cut -d '"' -f4 | xargs curl -L -o vimongo
chmod +x vimongo
sudo mv vimongo /usr/bin
```

### Using wget

```bash
wget -O - https://api.github.com/repos/kopecmaciej/vimongo/releases/latest | jq -r '.assets[0].browser_download_url' | xargs wget
chmod +x vimongo
sudo mv vimongo /usr/bin
```

### Using Go

```bash
git clone git@github.com:kopecmaciej/vimongo.git
cd vimongo
make
```

## Usage

After installing Vi Mongo, you can run it by typing `vimongo` in your terminal.

```bash
vimongo
```

In any moment you can press `Ctrl + H` to see help page with all available
shortcuts. Resizing terminal while running Vi Mongo should work fine, but if you
encounter any issues, please let me know.

All configuration files should be stored in `~/.config/vimongo` directory, but
it depends on the system settings as env `XDG_CONFIG_HOME` can be set to
different directories (more information here:
[XDG Base Directory](https://github.com/adrg/xdg?tab=readme-ov-file#xdg-base-directory))

## List of features to be implemented

- [x] Query History
- [x] Switching between multiple Connections
- [x] Help page
- [x] Improve Content by adding other possibilities of viewing
- [x] Move autocomplete keys to json file, so that it can be easily modified
- [ ] Hash passwords in config
- [ ] Multiple styles
- [ ] Managing Indexes
- [ ] Aggregation Pipeline
- [ ] Exporting/Importing Documents

## Issues

- [x] Searching collection on databases not expanding tree
- [x] No clearing history
- [x] No view with shortcuts
- [ ] Header not updated after changing database/collection
- [ ] Content not updated after editing from picker
- [ ] Performance issue while loading large bson files
