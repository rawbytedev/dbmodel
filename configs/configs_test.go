package configs_test

import (
	"dbmodel/configs"
	"fmt"
	"testing"
)

func TestEmptyConfigsComp(t *testing.T) {
	a := configs.StoreConfig{}
	if a.BadgerConfigs != nil {
		fmt.Print("True")
	} else {
		fmt.Print("False")
	}
}
