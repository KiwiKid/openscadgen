package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// SetupMCPServer sets up MCP (Model Context Protocol) endpoints
func SetupMCPServer() {
	// MCP endpoints
	http.HandleFunc("/v1/metadata", handleMCPMetadata)
	http.HandleFunc("/v1/resources", handleMCPResources)
	http.HandleFunc("/v1/resource/", handleMCPResource)
	http.HandleFunc("/v1/tools", handleMCPTools)
	http.HandleFunc("/v1/tool/", handleMCPTool)
}

// handleMCPMetadata handles the MCP metadata endpoint
func handleMCPMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":         "openscadgen",
		"version":      pkg.GetVersion().OpenSCADGen,
		"capabilities": []string{"resources", "tools"},
	})
}

// handleMCPResources handles the MCP resources endpoint
func handleMCPResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{"id": "cube-config", "type": "openscadgen-config", "description": "Cube config for openscadgen"},
	})
}

// handleMCPResource handles individual MCP resource requests
func handleMCPResource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Path[len("/v1/resource/"):]
	configPath := r.URL.Query().Get("config")
	if configPath == "" {
		configPath = "bols2otropolis/cube/config.toml"
	}
	config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configPath, Server: true})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var instanceSummaries []map[string]interface{}
	for _, inst := range config.Design.ConfiguredInstanceConfig {
		instanceSummaries = append(instanceSummaries, map[string]interface{}{
			"name":        inst.Name,
			"params":      inst.Params,
			"description": inst.Description,
		})
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"config_path": config.ConfigFile,
		"instances":   instanceSummaries,
		"raw_config":  config.RawConfigFile,
	})
}

// handleMCPTools handles the MCP tools endpoint
func handleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{"id": "list_instances", "description": "List all configured instances"},
		{"id": "process_instance", "description": "Process a single instance (POST: {\"name\":...})"},
		{"id": "process_all", "description": "Process all instances"},
	})
}

// handleMCPTool handles individual MCP tool invocations
func handleMCPTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !((r.Method == "POST") && (len(r.URL.Path) > len("/v1/tool/") && r.URL.Path[len(r.URL.Path)-len("/invoke"):] == "/invoke")) {
		http.NotFound(w, r)
		return
	}
	id := r.URL.Path[len("/v1/tool/") : len(r.URL.Path)-len("/invoke")]
	var req struct {
		Config string `json:"config"`
		Name   string `json:"name"`
	}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &req)
	configPath := req.Config
	if configPath == "" {
		configPath = "bols2otropolis/cube/config.toml"
	}
	config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configPath, Server: true})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	switch id {
	case "list_instances":
		var instanceNames []string
		for _, inst := range config.Design.ConfiguredInstanceConfig {
			instanceNames = append(instanceNames, inst.Name)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": instanceNames,
		})
	case "process_instance":
		if req.Name == "" {
			http.Error(w, "Missing or invalid instance name", 400)
			return
		}
		config.RegexPattern = req.Name
		result, err := pkg.Process(config, &pkg.NoopProgress{}, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "Processed instance",
			"output": result,
		})
	case "process_all":
		result, err := pkg.Process(config, &pkg.NoopProgress{}, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "Processed all instances",
			"output": result,
		})
	default:
		http.NotFound(w, r)
	}
}
