package project

// Re-export types for backwards compatibility
// Main types are in internal/types package

import "github.com/kraitsura/conduit/internal/types"

type Project = types.Project
type Agent = types.Agent
type Activity = types.Activity
type ProjectStats = types.ProjectStats
type Commit = types.Commit
type DaemonStatus = types.DaemonStatus

// Activity type constants
const (
	ActivityAgentStart = types.ActivityAgentStart
	ActivityAgentStop  = types.ActivityAgentStop
	ActivityCommit     = types.ActivityCommit
	ActivityBranch     = types.ActivityBranch
	ActivityFileChange = types.ActivityFileChange
)
