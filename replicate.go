package carp

type UserAttibutes map[string][]string

type UserReplicator func(username string, attributes UserAttibutes) error
