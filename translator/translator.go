package translator

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigMapRef struct {
	Name string `yaml:"name"`
}

type EnvFromEntry struct {
	ConfigMapRef ConfigMapRef `yaml:"configMapRef"`
}

type SealedSecrets struct {
	Scope     string            `yaml:"scope"`
	SecretRef map[string]string `yaml:"secretRef"`
}

type ExtEnvVarsFrom struct {
	Enabled       bool            `yaml:"enabled"`
	EnvFrom       []EnvFromEntry  `yaml:"envFrom,omitempty"`
	SealedSecrets *SealedSecrets  `yaml:"sealedSecrets,omitempty"`
}

type PortEntry struct {
	Name          string `yaml:"name"`
	ContainerPort int    `yaml:"containerPort"`
	ServicePort   int    `yaml:"servicePort"`
	Protocol      string `yaml:"protocol"`
}

type SecretKeyRef struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ValueFrom struct {
	SecretKeyRef SecretKeyRef `yaml:"secretKeyRef"`
}

type ExtEnvVarEntry struct {
	Name      string    `yaml:"name"`
	ValueFrom ValueFrom `yaml:"valueFrom"`
}

type ResourceLimits struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type ResourceRequests struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type Resources struct {
	Requests ResourceRequests `yaml:"requests"`
	Limits   ResourceLimits   `yaml:"limits"`
}

type Deploy struct {
	ReplicaCount     int              `yaml:"replicaCount"`
	ImagePullSecrets []map[string]string `yaml:"imagePullSecrets"`
	Image            Image            `yaml:"image"`
	EnvVars          map[string]any   `yaml:"envVars"`
	ExtEnvVars       []ExtEnvVarEntry `yaml:"extEnvVars,omitempty"`
	Resources        Resources        `yaml:"resources"`
}

type Image struct {
	Repository string      `yaml:"repository"`
	Tag        string      `yaml:"tag"`
	Ports      []PortEntry `yaml:"ports"`
}

type IngressPath struct {
	Path     string `yaml:"path"`
	PathType string `yaml:"pathType"`
}

type IngressHost struct {
	Host  string        `yaml:"host,omitempty"`
	Paths []IngressPath `yaml:"paths"`
}

type IngressEntry struct {
	Name             string        `yaml:"name"`
	Enabled          bool          `yaml:"enabled"`
	IngressClassName string        `yaml:"ingressClassName"`
	Hosts            []IngressHost `yaml:"hosts"`
}

type Networking struct {
	Ingress []IngressEntry `yaml:"ingress"`
}

type WebappValues struct {
	NameOverride   string         `yaml:"nameOverride"`
	Cluster        string         `yaml:"cluster"`
	Environment    string         `yaml:"environment"`
	Team           string         `yaml:"team"`
	Project        string         `yaml:"project"`
	ExtEnvVarsFrom ExtEnvVarsFrom `yaml:"extEnvVarsFrom"`
	Deploy         Deploy         `yaml:"deploy"`
	Networking     Networking     `yaml:"networking"`
}

var defaultExcludeKeys = map[string]bool{
	"CONFIG_SERVER_ADDR":      true,
	"CONFIG_SERVER_ENABLED":   true,
	"CONFIG_DATABASE_ENABLED": true,
	"REDIS_SERVER":            true,
	"REDIS_PORT":              true,
	"REDIS_PASSWORD":          true,
	"REDIS_TIMEOUT":           true,
	"PROJECT_CACHE_TYPE":      true,
	"MONITOR_SERVER":          true,
	"MONITOR_ENABLED":         true,
	"MONITOR_USER":            true,
	"MONITOR_PASSWORD":        true,
	"ORACLE_SERVER":           true,
	"ORACLE_PORT":             true,
	"ORACLE_DATABASE":         true,
	"ORACLE_USER":             true,
	"ORACLE_PASSWORD":         true,
	"SEALED_ENV_VAR":          true,
	"ENV_VAR":                 true,
}

func TranslateValues(oldYamlStr string, cluster string, envOverride string, teamOverride string, javaOptionsOverride string, useCommonConfigmap bool) (string, string, error) {
	var oldData map[string]any
	if err := yaml.Unmarshal([]byte(oldYamlStr), &oldData); err != nil {
		return "", "", fmt.Errorf("error parsing old YAML: %w", err)
	}
	if len(oldData) == 0 {
		return "", "", fmt.Errorf("empty YAML document or invalid content")
	}

	// 1. Metadata Extraction
	nameOverride := "unknown-app"
	if v, ok := oldData["nameOverride"].(string); ok && v != "" {
		nameOverride = v
	} else if v, ok := oldData["fullnameOverride"].(string); ok && v != "" {
		nameOverride = v
	}
	nameOverride = strings.ToLower(strings.TrimSpace(nameOverride))

	environment := "develop"
	if envOverride != "" {
		environment = envOverride
	} else if v, ok := oldData["environment"].(string); ok && v != "" {
		environment = v
	}

	team := "middleware"
	if teamOverride != "" {
		team = teamOverride
	} else if v, ok := oldData["team"].(string); ok && v != "" {
		team = v
	}
	project := team

	// 2. Build Ingress Host and Paths
	ingressMap, _ := oldData["ingress"].(map[string]any)
	ingressEnabled := true
	if ie, ok := ingressMap["enabled"].(bool); ok {
		ingressEnabled = ie
	}

	var ingressPaths []string
	var oldHost string

	if hostsList, ok := ingressMap["hosts"].([]any); ok && len(hostsList) > 0 {
		if firstHost, ok := hostsList[0].(map[string]any); ok {
			if h, ok := firstHost["host"].(string); ok {
				oldHost = h
			}
			if pathsList, ok := firstHost["paths"].([]any); ok {
				for _, p := range pathsList {
					if pm, ok := p.(map[string]any); ok {
						if pathStr, ok := pm["path"].(string); ok {
							ingressPaths = append(ingressPaths, pathStr)
						}
					} else if ps, ok := p.(string); ok {
						ingressPaths = append(ingressPaths, ps)
					}
				}
			}
		}
	}

	if oldHost == "" || strings.Contains(oldHost, "<name>") {
		envShort := "dev"
		if strings.Contains(environment, "stg") || strings.Contains(environment, "staging") {
			envShort = "stg"
		} else if strings.Contains(environment, "release") || strings.Contains(environment, "prep") {
			envShort = "prep"
		} else if strings.Contains(environment, "prod") || strings.Contains(environment, "master") {
			envShort = "prod"
		}
		if envShort == "prod" {
			oldHost = "api.bg.com.bo"
		} else {
			oldHost = fmt.Sprintf("api.%s.bg.com.bo", envShort)
		}
	}

	if len(ingressPaths) == 0 {
		ingressPaths = []string{"/" + nameOverride + "/"}
	}

	// 3. Handle Environment Variables (Configmap)
	configmapMap, _ := oldData["configmap"].(map[string]any)
	envVars := make(map[string]any)
	hasOracle := false

	for k, v := range configmapMap {
		if strings.Contains(k, "ORACLE_") {
			hasOracle = true
		}
		if !defaultExcludeKeys[k] {
			envVars[k] = v
		}
	}

	if _, exists := envVars["JAVA_TOOL_OPTIONS"]; !exists && javaOptionsOverride != "" {
		envVars["JAVA_TOOL_OPTIONS"] = javaOptionsOverride
	}

	// 4. Handle Sealed Secrets
	var sealedSecrets map[string]string
	sealedMap, _ := oldData["sealed"].(map[string]any)
	if sealedMap != nil {
		if sealedDataMap, ok := sealedMap["data"].(map[string]any); ok {
			sealedSecrets = make(map[string]string)
			for k, v := range sealedDataMap {
				if vStr, ok := v.(string); ok && vStr != "" {
					sealedSecrets[k] = vStr
					if strings.Contains(k, "ORACLE_") {
						hasOracle = true
					}
				}
			}
		}
	}

	// 5. extEnvVarsFrom
	extEnvVarsFrom := ExtEnvVarsFrom{
		Enabled: false,
	}
	if useCommonConfigmap || len(sealedSecrets) > 0 {
		extEnvVarsFrom.Enabled = true
		if useCommonConfigmap {
			extEnvVarsFrom.EnvFrom = []EnvFromEntry{
				{
					ConfigMapRef: ConfigMapRef{
						Name: "commun",
					},
				},
			}
		}
		if len(sealedSecrets) > 0 {
			extEnvVarsFrom.SealedSecrets = &SealedSecrets{
				Scope:     "strict",
				SecretRef: sealedSecrets,
			}
		}
	}

	// 6. Deploy setup
	replicaCount := 1
	if rc, ok := oldData["replicaCount"].(int); ok {
		if rc > 0 {
			replicaCount = rc
		}
	} else if rcFloat, ok := oldData["replicaCount"].(float64); ok {
		if int(rcFloat) > 0 {
			replicaCount = int(rcFloat)
		}
	}

	imageMap, _ := oldData["image"].(map[string]any)
	imageRepo := ""
	if ir, ok := imageMap["repository"].(string); ok && ir != "" {
		imageRepo = ir
	} else {
		imageRepo = fmt.Sprintf("bgdevopsreg.azurecr.io/%s/%s", team, nameOverride)
	}

	imageTag := "latest"
	if it, ok := imageMap["tag"].(string); ok && it != "" {
		imageTag = it
	} else if itFloat, ok := imageMap["tag"].(float64); ok {
		imageTag = fmt.Sprintf("%g", itFloat)
	}

	serviceMap, _ := oldData["service"].(map[string]any)
	containerPort := 8080
	if cp, ok := serviceMap["targetPort"].(int); ok {
		containerPort = cp
	} else if cp, ok := serviceMap["port"].(int); ok {
		containerPort = cp
	} else if cpFloat, ok := serviceMap["targetPort"].(float64); ok {
		containerPort = int(cpFloat)
	} else if cpFloat, ok := serviceMap["port"].(float64); ok {
		containerPort = int(cpFloat)
	}

	servicePort := 8080
	if sp, ok := serviceMap["port"].(int); ok {
		servicePort = sp
	} else if spFloat, ok := serviceMap["port"].(float64); ok {
		servicePort = int(spFloat)
	}

	protocol := "TCP"
	if pr, ok := serviceMap["protocol"].(string); ok && pr != "" {
		protocol = pr
	}

	portName := "http"
	if pn, ok := serviceMap["portName"].(string); ok && pn != "" {
		portName = pn
	}

	deployImage := Image{
		Repository: imageRepo,
		Tag:        imageTag,
		Ports: []PortEntry{
			{
				Name:          portName,
				ContainerPort: containerPort,
				ServicePort:   servicePort,
				Protocol:      protocol,
			},
		},
	}

	var imagePullSecrets []map[string]string
	if ips, ok := oldData["imagePullSecrets"].([]any); ok && len(ips) > 0 {
		for _, secret := range ips {
			if sMap, ok := secret.(map[string]any); ok {
				secEntry := make(map[string]string)
				for sk, sv := range sMap {
					if svStr, ok := sv.(string); ok {
						secEntry[sk] = svStr
					}
				}
				imagePullSecrets = append(imagePullSecrets, secEntry)
			}
		}
	}
	if len(imagePullSecrets) == 0 {
		imagePullSecrets = []map[string]string{
			{"name": "azure-container-registry"},
		}
	}

	resourcesMap, _ := oldData["resources"].(map[string]any)
	requestsMap, _ := resourcesMap["requests"].(map[string]any)
	limitsMap, _ := resourcesMap["limits"].(map[string]any)

	cpuReq := "250m"
	if cr, ok := requestsMap["cpu"].(string); ok && cr != "" {
		cpuReq = cr
	}
	memReq := "512Mi"
	if mr, ok := requestsMap["memory"].(string); ok && mr != "" {
		memReq = mr
	}
	cpuLim := "500m"
	if cl, ok := limitsMap["cpu"].(string); ok && cl != "" {
		cpuLim = cl
	}
	memLim := "1Gi"
	if ml, ok := limitsMap["memory"].(string); ok && ml != "" {
		memLim = ml
	}

	deploy := Deploy{
		ReplicaCount:     replicaCount,
		ImagePullSecrets: imagePullSecrets,
		Image:            deployImage,
		EnvVars:          envVars,
		Resources: Resources{
			Requests: ResourceRequests{
				CPU:    cpuReq,
				Memory: memReq,
			},
			Limits: ResourceLimits{
				CPU:    cpuLim,
				Memory: memLim,
			},
		},
	}

	if hasOracle && len(sealedSecrets) == 0 {
		deploy.ExtEnvVars = []ExtEnvVarEntry{
			{Name: "ORACLE_SERVER", ValueFrom: ValueFrom{SecretKeyRef: SecretKeyRef{Name: "oracle-conn", Key: "ORA_HOST"}}},
			{Name: "ORACLE_PORT", ValueFrom: ValueFrom{SecretKeyRef: SecretKeyRef{Name: "oracle-conn", Key: "ORA_PORT"}}},
			{Name: "ORACLE_DATABASE", ValueFrom: ValueFrom{SecretKeyRef: SecretKeyRef{Name: "oracle-conn", Key: "ORA_DATABASE"}}},
			{Name: "ORACLE_USER", ValueFrom: ValueFrom{SecretKeyRef: SecretKeyRef{Name: "oracle-conn", Key: "ORA_USER_MIDDLEWARE"}}},
			{Name: "ORACLE_PASSWORD", ValueFrom: ValueFrom{SecretKeyRef: SecretKeyRef{Name: "oracle-conn", Key: "ORA_PASS_MIDDLEWARE"}}},
		}
	}

	// 7. Networking setup
	var ingressHosts []IngressHost
	var ingressPathsObj []IngressPath
	for _, p := range ingressPaths {
		ingressPathsObj = append(ingressPathsObj, IngressPath{
			Path:     p,
			PathType: "ImplementationSpecific",
		})
	}

	ingressHosts = append(ingressHosts, IngressHost{
		Host:  oldHost,
		Paths: ingressPathsObj,
	})

	ingressHosts = append(ingressHosts, IngressHost{
		Paths: ingressPathsObj,
	})

	networking := Networking{
		Ingress: []IngressEntry{
			{
				Name:             "default",
				Enabled:          ingressEnabled,
				IngressClassName: "kong",
				Hosts:            ingressHosts,
			},
		},
	}

	webappValues := WebappValues{
		NameOverride:   nameOverride,
		Cluster:        cluster,
		Environment:    environment,
		Team:           team,
		Project:        project,
		ExtEnvVarsFrom: extEnvVarsFrom,
		Deploy:         deploy,
		Networking:     networking,
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2) // Set indent to standard 2-spaces!
	if err := enc.Encode(webappValues); err != nil {
		return "", "", fmt.Errorf("error serializing translated YAML: %w", err)
	}

	yamlHeader := "########################################\n## CONFIG | Web server values\n########################################\n"
	finalYaml := yamlHeader + unwrapYaml(buf.String())

	targetPath := fmt.Sprintf("charts/webapp/values/%s/%s/%s/%s.yaml", cluster, environment, team, nameOverride)

	return finalYaml, targetPath, nil
}

func unwrapYaml(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}
		hasHash := strings.HasPrefix(trimmed, "#")
		hasColon := strings.Contains(line, ":")
		hasDash := strings.HasPrefix(trimmed, "-")
		isCont := !hasHash && !hasColon && !hasDash
		if isCont {
			if len(result) > 0 {
				result[len(result)-1] = result[len(result)-1] + trimmed
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
