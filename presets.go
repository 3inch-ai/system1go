package system1go

// Capabilities declares the implemented common surface, not all server features.
type Capabilities struct {
	MinChoices, MaxChoices, MinLevels, MaxLevels, MaxQuestions int
	BooleanChoices, StructuredInstructions, Refusal            bool
}

// Capabilities returns a copy of this client's declared protocol limits.
func (c *Client) Capabilities() Capabilities {
	v := Capabilities{MinChoices: 1, MaxChoices: 255, MinLevels: 1, MaxLevels: 10, StructuredInstructions: true}
	if c.config.Protocol == OpenAIDecisions {
		return Capabilities{MinChoices: 2, MaxChoices: 255, MinLevels: 2, MaxLevels: 10, MaxQuestions: 400, BooleanChoices: true, Refusal: true}
	}
	if c.config.MaxQuestions > 0 {
		v.MaxQuestions = c.config.MaxQuestions
	}
	if c.config.MaxChoices > 0 {
		v.MaxChoices = c.config.MaxChoices
	}
	if c.config.MaxLevels > 0 {
		v.MaxLevels = c.config.MaxLevels
	}
	return v
}

// Jev defaults to the official TypeSafe endpoint and mutable jev-latest alias.
func Jev(c Config) (*Client, error) {
	c.Protocol = SystemOne
	c.Provider = "typesafe"
	if c.Model == "" {
		c.Model = "jev-latest"
	}
	if c.BaseURL == "" {
		c.BaseURL = "https://api.typesafe.ai"
	}
	return NewClient(c)
}

// OpenAI defaults to the official Decisions endpoint; this is not chat completions.
func OpenAI(c Config) (*Client, error) {
	c.Protocol = OpenAIDecisions
	c.Provider = "openai"
	if c.Model == "" {
		c.Model = "gpt-6-luna"
	}
	if c.BaseURL == "" {
		c.BaseURL = "https://api.openai.com"
	}
	return NewClient(c)
}

// H2O connects to an existing native shim. It never starts or downloads models.
func H2O(c Config) (*Client, error) {
	c.Protocol = SystemOne
	c.Provider = "h2o"
	if c.Model == "" {
		c.Model = "h2oai/h2o-lightning-4b"
	}
	if c.BaseURL == "" {
		c.BaseURL = "http://127.0.0.1:8741"
	}
	return NewClient(c)
}
