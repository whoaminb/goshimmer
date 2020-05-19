package shutdown

const (
	PriorityDatabase = iota
	PriorityWaspConn
	PriorityFPC
	PriorityTangle
	PriorityMissingMessagesMonitoring
	PriorityRemoteLog
	PriorityAnalysis
	PriorityMetrics
	PriorityAutopeering
	PriorityGossip
	PriorityWebAPI
	PriorityDashboard
	PrioritySynchronization
	PriorityBootstrap
	PrioritySpammer
	PriorityBadgerGarbageCollection
)
