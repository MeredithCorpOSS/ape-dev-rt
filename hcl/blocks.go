package hcl

import (
	"fmt"
	"log"
	"math"
	"sort"
)

type HclConfig struct {
	DeploymentState *DeploymentState
	RemoteState     *RemoteState
}

type DeploymentState struct {
	cfg []map[string]interface{}
}

type RemoteState struct {
	Backend string
	Config  map[string]string
}

func (ds *DeploymentState) Iterator() []map[string]interface{} {
	return ds.cfg
}

// block name => allowed occurence in cfg
var supportedBlocks = map[string]int{
	"deployment_state": math.MaxInt32,
	"remote_state":     1,
}

func parseBlock(hclConfig *HclConfig, blockKey string, cfgs []map[string]interface{}) error {
	allowedOccurence, ok := supportedBlocks[blockKey]
	if !ok {
		return fmt.Errorf("Unrecognised config block (%q), supported: %q",
			blockKey, getSupportedBlockNames(supportedBlocks))
	}

	occurence := len(cfgs)
	if occurence > allowedOccurence {
		return fmt.Errorf("Found %d occurences of %q. %q can only occur %d x times in the config.",
			occurence, blockKey, blockKey, allowedOccurence)
	}

	// Block specific
	if blockKey == "deployment_state" {
		hclConfig.DeploymentState = &DeploymentState{cfgs}
		return nil
	}

	if blockKey == "remote_state" {
		if len(cfgs) < 0 {
			return fmt.Errorf("No configuration provided for %q", blockKey)
		}

		var backend string
		var config = make(map[string]string, 0)
		for _, c := range cfgs[0] {
			cfg := c.([]map[string]interface{})
			log.Printf("CFG: %#v", cfg)
			if b, ok := cfg[0]["backend"]; ok {
				backend = b.(string)
			} else {
				return fmt.Errorf("Missing 'backend' field in %q", blockKey)
			}
			if c, ok := cfg[0]["config"]; ok {
				_cfg := c.([]map[string]interface{})
				for k, v := range _cfg[0] {
					config[k] = v.(string)
				}
			} else {
				return fmt.Errorf("Missing 'config' field in %q", blockKey)
			}

			hclConfig.RemoteState = &RemoteState{
				Backend: backend,
				Config:  config,
			}
		}
		return nil
	}

	return fmt.Errorf("Unable to parse block %q - no handler", blockKey)
}

func getSupportedBlockNames(mapping map[string]int) []string {
	var names = make([]string, len(mapping), len(mapping))
	i := 0
	for k, _ := range mapping {
		names[i] = k
		i++
	}
	sort.Strings(names)
	return names
}
