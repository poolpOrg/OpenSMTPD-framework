package main

import (
	"github.com/poolpOrg/OpenSMTPD-framework/table"
)

func main() {
	table.Init()

	table.OnUpdate(func() {
	})

	table.OnCheck(table.K_ALIAS, func(key string) (bool, error) {
		return true, nil
	})

	table.OnLookup(table.K_ALIAS, func(key string) (string, error) {
		return "", nil
	})

	table.OnFetch(table.K_ALIAS, func() (string, error) {
		return "", nil
	})

	table.Dispatch()
}
