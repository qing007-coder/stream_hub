package utils

import (
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

func CreateID() string {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	id := node.Generate()

	return id.String()
}

func CreateUUID() string {
	return uuid.New().String()
}
