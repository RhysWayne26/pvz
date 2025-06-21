package queries

import _ "embed"

// SaveOrderSQL contains the embedded SQL query for saving an order into the database.
//
//go:embed order/save.sql
var SaveOrderSQL string

// LoadOrderSQL contains the embedded SQL query for loading an order from the database by its ID.
//
//go:embed order/load.sql
var LoadOrderSQL string

// SoftDeleteOrderSQL holds the embedded SQL script for performing a soft delete operation on an order record.
//
//go:embed order/soft_delete.sql
var SoftDeleteOrderSQL string

// SaveHistoryEntrySQL contains the embedded SQL query used to save a history entry record into the database.
//
//go:embed history/save_entry.sql
var SaveHistoryEntrySQL string
