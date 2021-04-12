package hcl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/hashicorp/hcl"
)

const oldConfigFilename = "deployment-state.hcl.tpl"
const ConfigFilename = "rt.hcl.tpl"

type TemplateVariables struct {
	Environment  string
	AwsAccountId string
}

func LoadConfigFromPath(env, awsAccId, cfgPath string) (*HclConfig, string, error) {
	fileName := ConfigFilename

Load:
	// If it's a directory, use the default hcl filename
	filePath := cfgPath
	fileInfo, err := os.Stat(cfgPath)
	if err != nil || fileInfo.IsDir() {
		filePath = path.Join(cfgPath, fileName)
	}

	var configReader io.Reader
	log.Printf("[DEBUG] Trying to load config from %s", filePath)
	configReader, err = os.Open(filePath)
	if err != nil {
		// Try loading config from old filename if current doesn't exist
		if os.IsNotExist(err) && fileName != oldConfigFilename {
			fileName = oldConfigFilename
			goto Load
		} else {
			return nil, filePath, err
		}
	}

	cfg, err := ParseConfig(configReader, TemplateVariables{env, awsAccId})
	if err != nil {
		return nil, filePath, fmt.Errorf("Failed to load config from %q: %s", filePath, err)
	}

	return cfg, filePath, nil
}

// TODO: Make templating optional
func ParseConfig(configReader io.Reader, variables interface{}) (*HclConfig, error) {
	m, err := DecodeHCLFromTemplate(configReader, variables)
	if err != nil {
		return nil, err
	}

	hclConfig := &HclConfig{}

	for blockKey, blockCfg := range m {
		if cfgs, ok := blockCfg.([]map[string]interface{}); ok {
			err := parseBlock(hclConfig, blockKey, cfgs)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Unable to convert configuration of %s: %q", blockKey, blockCfg)
		}
	}

	return hclConfig, nil
}

func DecodeHCLFromTemplate(templateReader io.Reader, variables interface{}) (map[string]interface{}, error) {
	b := bytes.Buffer{}
	_, err := b.ReadFrom(templateReader)
	if err != nil {
		return nil, err
	}
	hclTemplateInBytes := b.Bytes()

	// Parse the template
	t := template.Must(template.New("-").Parse(string(hclTemplateInBytes)))

	var processedTplString bytes.Buffer
	err = t.Execute(&processedTplString, variables)
	if err != nil {
		return nil, err
	}

	// And now let's decode the HCL
	var configs map[string]interface{}
	err = hcl.Decode(&configs, processedTplString.String())
	if err != nil {
		return nil, fmt.Errorf("Unable to decode HCL: %s", err)
	}

	return configs, nil
}
