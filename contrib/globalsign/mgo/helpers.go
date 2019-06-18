package mgo

import (
	"fmt"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
)

func setTagsForCommand(tags map[string]string, command string) map[string]string {
	if tags == nil {
		tags = make(map[string]string)
	}

	tags[ext.ResourceName] = fmt.Sprintf("mongo.%s", command)
	if dbName, ok := tags[ext.DBName]; ok {
		tags[ext.DBStatement] = fmt.Sprintf("%s %s", command, dbName)
	} else {
		// DBName not set in tags
	}
	tags[ext.ResourceName] = fmt.Sprintf("mongo.%s", command)

	return tags
}
