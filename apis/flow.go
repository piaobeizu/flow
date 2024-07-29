/*
 @Version : 1.0
 @Author  : steven.wong
 @Email   : 'wangxk1991@gamil.com'
 @Time    : 2023/04/01 22:13:34
 Desc     :
*/

package apis

import (
	"encoding/json"

	"github.com/jellydator/validation"
)

// APIVersion is the current api version
const APIVersion = "flow.github.com/v1beta1"

// Hooks define a list of hooks such as hooks["apply"]["before"] = ["ls -al", "rm foo.txt"]
type hooks map[string]map[string][]string

// ForActionAndStage return hooks for given action and stage

type Metadata struct {
	Name    string
	Version string
}

type Spec struct {
	Config any
}

type FlowConfig struct {
	APIVersion string
	Kind       string
	Metadata   *Metadata
	Hooks      hooks
	Spec       *Spec
}

func NewFlowConfig(name, version string) *FlowConfig {
	return &FlowConfig{
		APIVersion: APIVersion,
		Kind:       "flow",
		Metadata: &Metadata{
			Name:    name,
			Version: version,
		},
		Spec: &Spec{},
	}
}

func (c *FlowConfig) SetSpec(config any) *FlowConfig {
	if c.Spec == nil {
		c.Spec = &Spec{}
	}
	c.Spec.Config = config
	return c
}

func (c *FlowConfig) SetHooks(hooks map[string]map[string][]string) *FlowConfig {
	c.Hooks = hooks
	return c
}

func (c *FlowConfig) ForActionAndStage(action, stage string) []string {
	if len(c.Hooks[action]) > 0 {
		return c.Hooks[action][stage]
	}
	return nil
}

// UnmarshalYAML sets in some sane defaults when unmarshaling the data from yaml
func (c *FlowConfig) Unmarshal(config any) error {
	str, err := json.Marshal(c.Spec.Config)
	if err != nil {
		return err
	}
	return json.Unmarshal(str, config)
}

// UnmarshalYAML sets in some sane defaults when unmarshaling the data from yaml
func (c *FlowConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Metadata = &Metadata{
		Name: "flow",
	}
	c.Spec = &Spec{}

	type apolloConfig FlowConfig
	yc := (*apolloConfig)(c)

	if err := unmarshal(yc); err != nil {
		return err
	}
	// log.Printf("test UnmarshalYAML for cluster: %v", c)
	return nil
}

// Validate performs a configuration sanity check
func (c *FlowConfig) Validate() error {
	validation.ErrorTag = "yaml"
	return validation.ValidateStruct(c,
		validation.Field(&c.APIVersion, validation.Required, validation.In(APIVersion).Error("must equal "+APIVersion)),
		validation.Field(&c.Kind, validation.Required, validation.
			In("flow", "Flow").Error("must equal Cluster")),
		validation.Field(&c.Spec),
	)
}
