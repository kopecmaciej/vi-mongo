<p align="center"><img src="./assets/mongui_header.png"></p>

# MongUI: Terminal User Interface for MongoDB

### Managing connections

<p align="center"><img src="./assets/manage_connections.gif"></p>

### Working with documents

<p align="center"><img src="./assets/working_with_docs.gif"></p>

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

## Installation

For now, Mongui is available only by having Go installed on your machine. To
install Mongui, simply run the following command:

```
git clone git@github.com:kopecmaciej/mongui.git
cd mongui
make
```

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

- [x] Searching collection on sidebar not expanding tree
- [ ] No clearing history
- [ ] Header not updated after changing database/collection
- [ ] No view with shortcuts
- [ ] Content not updated after editing from picker
- [ ] Performance issue while loading large bson files
