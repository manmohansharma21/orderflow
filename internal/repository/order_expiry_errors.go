package repository

import "errors"

var ErrOrderExpiryInventoryConflict = errors.New("reserved inventory cannot be released while expiring an order")
