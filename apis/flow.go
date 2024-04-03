/*
 @Version : 1.0
 @Author  : steven.wong
 @Email   : 'wangxk1991@gamil.com'
 @Time    : 2023/04/01 22:13:34
 Desc     :
*/

package apis

import (
	"github.com/jellydator/validation"
)

// APIVersion is the current api version
const APIVersion = "flow.github.com/v1beta1"

type Metadata struct {
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
}

type Spec struct {
	Hooks *Hooks `yaml:"hooks,omitempty"`
}

type FlowConfig struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *Spec     `yaml:"spec"`
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
