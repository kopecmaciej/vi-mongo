# Mongui: Terminal User Interface for MongoDB

## WORK IN PROGRESS

## Overview

**Mongui** is an intuitive Terminal User Interface (TUI) application, crafted
with Go, designed to streamline and simplify the management of MongoDB
databases. Emphasizing ease of use without sacrificing functionality, Mongui
offers a user-friendly command-line experience for database administrators and
developers alike.

## Features

- **Intuitive Navigation**: Mongui's simple, intuitive interface makes it easy
  to navigate and manage your MongoDB databases.
- **Managing Documents**: Mongui allows you to view, create, update, duplicate
  and delete documents in your databases with ease.

## Installation

## Usage

Launch Mongui from the terminal:

`mongui`

Follow on-screen instructions for navigating and managing your MongoDB
databases.

## List of features to be implemented

- [x] Query History
- [ ] Switching between multiple Connections
- [ ] Improve Content by adding other possibilities of viewing
- [ ] Multiple styles
- [ ] Managing Indexes
- [ ] Aggregation Pipeline
- [ ] Exporting/Importing Documents
- [ ] Move autocomplete keys to json file, so that it can be easily modified

## Issues

- [ ] Searching collection on sidebar not expanding tree
- [ ] No clearing history
- [ ] Header not updated after changing database/collection
- [ ] No view with shortcuts
- [ ] Content not updated after editing from picker
- [ ] Performance issue while loading large bson files
