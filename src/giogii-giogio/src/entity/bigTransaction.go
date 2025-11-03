package entity

type BigTransaction struct {
	LockCount       *int64
	ProcesslistId   *int64
	ProcesslistUser string
	ProcesslistHost string
	ThreadId        *int64
	SqlText         string
}
