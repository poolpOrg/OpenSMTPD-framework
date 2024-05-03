package main

import (
	"time"

	"github.com/poolpOrg/OpenSMTPD-framework/table"
)

func main() {
	table.Init()

	table.OnUpdate(func() error {
		return nil
	})

	table.OnCheck(table.K_ALIAS, func(timestamp time.Time, table string, key string) (bool, error) {
		return true, nil
	})

	table.OnLookup(table.K_ALIAS, func(timestamp time.Time, table string, key string) (string, error) {
		return "", nil
	})

	table.OnFetch(table.K_ALIAS, func(timestamp time.Time, table string) (string, error) {
		return "", nil
	})

	table.Dispatch()
}
