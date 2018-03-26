package main

type UserAttibutes map[string][]string

type UserReplicator func(username string, attributes UserAttibutes) error
