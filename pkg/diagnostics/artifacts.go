package diagnostics

// ArtifactKind identifies a collected failure artifact type.
type ArtifactKind string

const (
	ArtifactAgentLogs ArtifactKind = "agent-logs"
	ArtifactEvents    ArtifactKind = "events"
	ArtifactDiff      ArtifactKind = "diff"
	ArtifactTimeline  ArtifactKind = "timeline"
)

// ArtifactDescriptor references a written diagnostic artifact.
type ArtifactDescriptor struct {
	Kind        ArtifactKind
	Path        string
	Description string
}

// Artifacts groups diagnostic outputs for a failed scenario.
type Artifacts struct {
	ScenarioName string
	Items        []ArtifactDescriptor
}
