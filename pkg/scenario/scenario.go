package scenario

// Scenario is a declarative test scenario loaded from YAML.
type Scenario struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Agents      AgentSet `yaml:"agents"`
	Setup       Setup    `yaml:"setup"`
	Trigger     *Trigger `yaml:"trigger,omitempty"`
	Expect      Expect   `yaml:"expect"`
}

// AgentSet lists agent identifiers participating in a scenario.
type AgentSet []string

// Setup describes initial cluster state before triggers and assertions.
type Setup struct {
	Manifests []string `yaml:"manifests"`
}
