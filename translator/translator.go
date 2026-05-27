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

// Argo CD Application Models
type ArgoApplicationMetadata struct {
	Name       string   `yaml:"name"`
	Namespace  string   `yaml:"namespace"`
	Finalizers []string `yaml:"finalizers,omitempty"`
}

type ArgoApplicationHelm struct {
	ValueFiles []string `yaml:"valueFiles"`
}

type ArgoApplicationSource struct {
	RepoURL        string               `yaml:"repoURL"`
	Path           string               `yaml:"path"`
	TargetRevision string               `yaml:"targetRevision"`
	Helm           *ArgoApplicationHelm `yaml:"helm,omitempty"`
}

type ArgoApplicationDestination struct {
	Server    string `yaml:"server"`
	Namespace string `yaml:"namespace"`
}

type ArgoApplicationSyncAutomated struct {
	Prune      bool `yaml:"prune"`
	SelfHeal   bool `yaml:"selfHeal"`
	AllowEmpty bool `yaml:"allowEmpty"`
}

type ArgoApplicationSyncPolicy struct {
	SyncOptions []string                      `yaml:"syncOptions,omitempty"`
	Automated   *ArgoApplicationSyncAutomated `yaml:"automated,omitempty"`
}

type ArgoApplicationSpec struct {
	Project     string                       `yaml:"project"`
	Sources     []ArgoApplicationSource      `yaml:"sources"`
	Destination ArgoApplicationDestination    `yaml:"destination"`
	SyncPolicy  *ArgoApplicationSyncPolicy   `yaml:"syncPolicy,omitempty"`
}

type ArgoApplication struct {
	APIVersion string                  `yaml:"apiVersion"`
	Kind       string                  `yaml:"kind"`
	Metadata   ArgoApplicationMetadata `yaml:"metadata"`
	Spec       ArgoApplicationSpec     `yaml:"spec"`
}

type LegacyApplication struct {
	Metadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Project string `yaml:"project"`
		Source  struct {
			RepoURL        string `yaml:"repoURL"`
			Path           string `yaml:"path"`
			TargetRevision string `yaml:"targetRevision"`
			Helm           struct {
				ValueFiles []string `yaml:"valueFiles"`
			} `yaml:"helm"`
		} `yaml:"source"`
		Destination struct {
			Server    string `yaml:"server"`
			Namespace string `yaml:"namespace"`
		} `yaml:"destination"`
	} `yaml:"spec"`
}

func TranslateValuesWithArgo(oldYamlStr string, cluster string, envOverride string, teamOverride string, javaOptionsOverride string, useCommonConfigmap bool) (string, string, string, string, string, error) {
	translatedValues, targetValuesPath, err := TranslateValues(oldYamlStr, cluster, envOverride, teamOverride, javaOptionsOverride, useCommonConfigmap)
	if err != nil {
		return "", "", "", "", "", err
	}

	var oldData map[string]any
	if err := yaml.Unmarshal([]byte(oldYamlStr), &oldData); err != nil {
		return "", "", "", "", "", err
	}

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

	argoApp, argoPath, err := TranslateArgoAppFromParams(nameOverride, team, environment, cluster, true)
	if err != nil {
		return "", "", "", "", "", err
	}

	legacyArgoPath := fmt.Sprintf("argocd/%s/%s/%s/%s.yaml", cluster, environment, team, nameOverride)

	return translatedValues, targetValuesPath, argoApp, argoPath, legacyArgoPath, nil
}

func TranslateArgoAppFromParams(appName string, team string, environment string, cluster string, syncPolicyEnabled bool) (string, string, error) {
	app := ArgoApplication{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Application",
		Metadata: ArgoApplicationMetadata{
			Name:       fmt.Sprintf("%s-%s", team, appName),
			Namespace:  "argocd",
			Finalizers: []string{"resources-finalizer.argocd.argoproj.io"},
		},
		Spec: ArgoApplicationSpec{
			Project: team,
			Sources: []ArgoApplicationSource{
				{
					RepoURL:        "https://BancoGanadero@dev.azure.com/BancoGanadero/BGA-DEVSECOPS/_git/charts",
					Path:           "./webapp",
					TargetRevision: "main",
					Helm: &ArgoApplicationHelm{
						ValueFiles: []string{
							fmt.Sprintf("values/%s/%s/%s/%s.yaml", cluster, environment, team, appName),
						},
					},
				},
			},
			Destination: ArgoApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: team,
			},
			SyncPolicy: &ArgoApplicationSyncPolicy{
				SyncOptions: []string{
					"Validate=false",
					"CreateNamespace=true",
					"PrunePropagationPolicy=foreground",
					"PruneLast=true",
					"RespectIgnoreDifferences=true",
				},
				Automated: &ArgoApplicationSyncAutomated{
					Prune:      true,
					SelfHeal:   true,
					AllowEmpty: false,
				},
			},
		},
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(app); err != nil {
		return "", "", fmt.Errorf("error serializing Argo App: %w", err)
	}

	yamlContent := buf.String()
	if !syncPolicyEnabled {
		yamlContent = commentOutSyncPolicy(yamlContent)
	}

	finalYaml := "---\n" + yamlContent
	targetPath := fmt.Sprintf("cluster/%s/%s/apps/%s/%s.yaml", cluster, environment, team, appName)

	return finalYaml, targetPath, nil
}

func TranslateArgoApp(legacyYamlStr string, cluster string, syncPolicyEnabled bool) (string, string, string, error) {
	strippedYaml := stripArgoComments(legacyYamlStr)

	var legacyApp LegacyApplication
	if err := yaml.Unmarshal([]byte(strippedYaml), &legacyApp); err != nil {
		return "", "", "", fmt.Errorf("error parsing legacy Argo Application: %w", err)
	}

	team := strings.ToLower(strings.TrimSpace(legacyApp.Spec.Destination.Namespace))
	if team == "" {
		team = strings.ToLower(strings.TrimSpace(legacyApp.Metadata.Namespace))
	}
	if team == "" {
		team = "middleware"
	}

	appName := ""
	pathParts := strings.Split(strings.Trim(legacyApp.Spec.Source.Path, "/"), "/")
	if len(pathParts) > 0 && pathParts[len(pathParts)-1] != "" {
		appName = pathParts[len(pathParts)-1]
	}
	if appName == "" {
		appName = legacyApp.Metadata.Name
		if team != "" && strings.HasPrefix(appName, team+"-") {
			appName = strings.TrimPrefix(appName, team+"-")
		}
	}
	appName = strings.ToLower(strings.TrimSpace(appName))

	environment := "develop"
	if len(legacyApp.Spec.Source.Helm.ValueFiles) > 0 {
		filename := legacyApp.Spec.Source.Helm.ValueFiles[0]
		cleaned := strings.TrimSuffix(strings.TrimPrefix(filename, "values-"), ".yaml")
		if cleaned != "" {
			environment = cleaned
		}
	}

	envLower := strings.ToLower(environment)
	if strings.Contains(envLower, "dev") {
		environment = "develop"
	} else if strings.Contains(envLower, "stg") || strings.Contains(envLower, "staging") {
		environment = "staging"
	} else if strings.Contains(envLower, "release") || strings.Contains(envLower, "prep") {
		environment = "release"
	} else if strings.Contains(envLower, "prod") || strings.Contains(envLower, "master") {
		environment = "production"
	}

	argoApp, argoPath, err := TranslateArgoAppFromParams(appName, team, environment, cluster, syncPolicyEnabled)
	if err != nil {
		return "", "", "", err
	}

	legacyArgoPath := fmt.Sprintf("argocd/%s/%s/%s/%s.yaml", cluster, environment, team, appName)

	return argoApp, argoPath, legacyArgoPath, nil
}

func stripArgoComments(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			stripped := strings.TrimPrefix(trimmed, "#")
			if strings.HasPrefix(stripped, " ") {
				stripped = strings.TrimPrefix(stripped, " ")
			}
			result = append(result, stripped)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func commentOutSyncPolicy(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	inSyncPolicy := false
	for i, line := range lines {
		if strings.HasPrefix(line, "  syncPolicy:") {
			inSyncPolicy = true
		}
		if inSyncPolicy {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(line, "  syncPolicy:") && !strings.HasPrefix(line, "    ") {
				inSyncPolicy = false
				continue
			}
			if line != "" {
				lines[i] = "  # " + strings.TrimPrefix(line, "  ")
			}
		}
	}
	return strings.Join(lines, "\n")
}
