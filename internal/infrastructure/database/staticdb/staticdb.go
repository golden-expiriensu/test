package staticdb

import (
	"fmt"

	"arkis_test/internal/domain/interfaces"
)

type StaticDb struct{}

var _ interfaces.Database = &StaticDb{}

func New() interfaces.Database {
	return &StaticDb{}
}

func (d *StaticDb) Get(input []byte) (string, error) {
	return fmt.Sprintf(`{"result": "%s"}`, string(input)), nil
}
