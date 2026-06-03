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

type ScalePolicy struct {
	Type          string `yaml:"type"`
	Value         int    `yaml:"value"`
	PeriodSeconds int    `yaml:"periodSeconds"`
}

type ScaleDown struct {
	Policies     []ScalePolicy `yaml:"policies,omitempty"`
	SelectPolicy string        `yaml:"selectPolicy,omitempty"`
}

type ScaleUp struct {
	Policies                   []ScalePolicy `yaml:"policies,omitempty"`
	SelectPolicy               string        `yaml:"selectPolicy,omitempty"`
	StabilizationWindowSeconds *int          `yaml:"stabilizationWindowSeconds,omitempty"`
}

type Behavior struct {
	ScaleDown ScaleDown `yaml:"scaleDown,omitempty"`
	ScaleUp   ScaleUp   `yaml:"scaleUp,omitempty"`
}

type Autoscaling struct {
	Enabled                           bool      `yaml:"enabled"`
	MinReplicas                       *int      `yaml:"minReplicas,omitempty"`
	MaxReplicas                       *int      `yaml:"maxReplicas,omitempty"`
	TargetCPUUtilizationPercentage    *int      `yaml:"targetCPUUtilizationPercentage,omitempty"`
	TargetMemoryUtilizationPercentage *int      `yaml:"targetMemoryUtilizationPercentage,omitempty"`
	Behavior                          *Behavior `yaml:"behavior,omitempty"`
}

type Storage struct {
	Enabled               bool           `yaml:"enabled"`
	PersistentVolume      map[string]any `yaml:"persistentVolume,omitempty"`
	PersistentVolumeClaim map[string]any `yaml:"persistentVolumeClaim,omitempty"`
}

type PortEntry struct {
	Name          string `yaml:"name"`
	ContainerPort int    `yaml:"containerPort"`
	ServicePort   int    `yaml:"servicePort"`
	Protocol      string `yaml:"protocol"`
}

type Image struct {
	Repository      string      `yaml:"repository"`
	Tag             string      `yaml:"tag"`
	ImagePullPolicy string      `yaml:"imagePullPolicy"`
	Ports           []PortEntry `yaml:"ports"`
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
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type ResourceRequests struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type Resources struct {
	Requests ResourceRequests `yaml:"requests"`
	Limits   ResourceLimits   `yaml:"limits"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}

type HostPath struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Volume struct {
	Name      string         `yaml:"name"`
	HostPath  *HostPath      `yaml:"hostPath,omitempty"`
	ConfigMap map[string]any `yaml:"configMap,omitempty"`
	Secret    map[string]any `yaml:"secret,omitempty"`
	NFS       map[string]any `yaml:"nfs,omitempty"`
}

type Deploy struct {
	ReplicaCount         int                 `yaml:"replicaCount"`
	ImagePullSecrets     []map[string]string `yaml:"imagePullSecrets,omitempty"`
	Image                Image               `yaml:"image"`
	EnvVars              map[string]any      `yaml:"envVars,omitempty"`
	ExtEnvVars           []ExtEnvVarEntry    `yaml:"extEnvVars,omitempty"`
	SecurityContext      map[string]any      `yaml:"securityContext,omitempty"`
	RevisionHistoryLimit *int                `yaml:"revisionHistoryLimit,omitempty"`
	Resources            Resources           `yaml:"resources"`
	VolumeMounts         []VolumeMount       `yaml:"volumeMounts,omitempty"`
	Volumes              []Volume            `yaml:"volumes,omitempty"`
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
	Name             string            `yaml:"name"`
	Enabled          bool              `yaml:"enabled"`
	IngressClassName string            `yaml:"ingressClassName,omitempty"`
	Annotations      map[string]string `yaml:"annotations,omitempty"`
	Hosts            []IngressHost     `yaml:"hosts"`
}

type KongPlugin struct {
	Name   string         `yaml:"name"`
	Plugin string         `yaml:"plugin"`
	Config map[string]any `yaml:"config"`
}

type Kong struct {
	Enabled *bool        `yaml:"enabled,omitempty"`
	Plugins []KongPlugin `yaml:"plugins,omitempty"`
}

type Networking struct {
	Ingress []IngressEntry `yaml:"ingress,omitempty"`
	Kong    *Kong          `yaml:"kong,omitempty"`
}

type WebappValues struct {
	NameOverride   string          `yaml:"nameOverride"`
	Cluster        string          `yaml:"cluster"`
	Environment    string          `yaml:"environment"`
	Team           string          `yaml:"team"`
	Project        string          `yaml:"project"`
	ExtEnvVarsFrom *ExtEnvVarsFrom `yaml:"extEnvVarsFrom,omitempty"`
	Autoscaling    *Autoscaling    `yaml:"autoscaling,omitempty"`
	Storage        *Storage        `yaml:"storage,omitempty"`
	Deploy         Deploy          `yaml:"deploy"`
	Networking     Networking      `yaml:"networking"`
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

type TranslationOptions struct {
	Cluster             string
	EnvOverride         string
	TeamOverride        string
	JavaOptionsOverride string
	UseCommonConfigmap  bool
	CpuRequest          string
	MemoryRequest       string
	CpuLimit            string
	MemoryLimit         string
	ForceResourceLimits bool
}

func TranslateValues(oldYamlStr string, cluster string, envOverride string, teamOverride string, javaOptionsOverride string, useCommonConfigmap bool) (string, string, error) {
	return TranslateValuesWithOptions(oldYamlStr, TranslationOptions{
		Cluster:             cluster,
		EnvOverride:         envOverride,
		TeamOverride:        teamOverride,
		JavaOptionsOverride: javaOptionsOverride,
		UseCommonConfigmap:  useCommonConfigmap,
	})
}

func TranslateValuesWithOptions(oldYamlStr string, opts TranslationOptions) (string, string, error) {
	if opts.CpuRequest == "" {
		opts.CpuRequest = "250m"
	}
	if opts.MemoryRequest == "" {
		opts.MemoryRequest = "512Mi"
	}
	if opts.CpuLimit == "" {
		opts.CpuLimit = "500m"
	}
	if opts.MemoryLimit == "" {
		opts.MemoryLimit = "1Gi"
	}

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
	if opts.EnvOverride != "" {
		environment = opts.EnvOverride
	} else if v, ok := oldData["environment"].(string); ok && v != "" {
		environment = v
	}

	team := "middleware"
	if opts.TeamOverride != "" {
		team = opts.TeamOverride
	} else if v, ok := oldData["team"].(string); ok && v != "" {
		team = v
	}
	project := team

	// 2. Autoscaling and Storage robust structures (nil if disabled to inherit from base values.yaml)
	var autoscaling *Autoscaling
	if autoMap, ok := oldData["autoscaling"].(map[string]any); ok {
		var a Autoscaling
		if autoBytes, err := yaml.Marshal(autoMap); err == nil {
			if err := yaml.Unmarshal(autoBytes, &a); err == nil {
				if a.Enabled {
					autoscaling = &a
				}
			}
		}
	}

	var storage *Storage
	if storeMap, ok := oldData["storage"].(map[string]any); ok {
		var s Storage
		if storeBytes, err := yaml.Marshal(storeMap); err == nil {
			if err := yaml.Unmarshal(storeBytes, &s); err == nil {
				if s.Enabled {
					storage = &s
				}
			}
		}
	}

	// 3. Handle Environment Variables (Configmap)
	configmapMap, _ := oldData["configmap"].(map[string]any)
	envVars := make(map[string]any)
	hasOracle := false

	for k, v := range configmapMap {
		if strings.Contains(k, "ORACLE_") {
			hasOracle = true
		}
		if !opts.UseCommonConfigmap || !defaultExcludeKeys[k] {
			envVars[k] = v
		}
	}

	if _, exists := envVars["JAVA_TOOL_OPTIONS"]; !exists && opts.JavaOptionsOverride != "" {
		envVars["JAVA_TOOL_OPTIONS"] = opts.JavaOptionsOverride
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

	// 5. extEnvVarsFrom (nil if disabled to inherit from base values.yaml)
	var extEnvVarsFrom *ExtEnvVarsFrom
	if opts.UseCommonConfigmap || len(sealedSecrets) > 0 {
		extEnvVarsFrom = &ExtEnvVarsFrom{
			Enabled: true,
		}
		if opts.UseCommonConfigmap {
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

	imagePullPolicy := "IfNotPresent"
	if ipp, ok := imageMap["imagePullPolicy"].(string); ok && ipp != "" {
		imagePullPolicy = ipp
	} else if ipp, ok := imageMap["pullPolicy"].(string); ok && ipp != "" {
		imagePullPolicy = ipp
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
		Repository:      imageRepo,
		Tag:             imageTag,
		ImagePullPolicy: imagePullPolicy,
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

	var revisionHistoryLimit *int
	if rhl, ok := oldData["revisionHistoryLimit"].(int); ok {
		if rhl != 5 {
			revisionHistoryLimit = &rhl
		}
	} else if rhlFloat, ok := oldData["revisionHistoryLimit"].(float64); ok {
		rhlInt := int(rhlFloat)
		if rhlInt != 5 {
			revisionHistoryLimit = &rhlInt
		}
	}

	resourcesMap, _ := oldData["resources"].(map[string]any)
	requestsMap, _ := resourcesMap["requests"].(map[string]any)
	limitsMap, _ := resourcesMap["limits"].(map[string]any)

	cpuReq := ""
	if cr, ok := requestsMap["cpu"].(string); ok && cr != "" {
		cpuReq = cr
	}
	memReq := ""
	if mr, ok := requestsMap["memory"].(string); ok && mr != "" {
		memReq = mr
	}
	cpuLim := ""
	if cl, ok := limitsMap["cpu"].(string); ok && cl != "" {
		cpuLim = cl
	}
	memLim := ""
	if ml, ok := limitsMap["memory"].(string); ok && ml != "" {
		memLim = ml
	}

	// Default fallbacks or strict overrides if specified
	if opts.ForceResourceLimits {
		cpuReq = opts.CpuRequest
		memReq = opts.MemoryRequest
		cpuLim = opts.CpuLimit
		memLim = opts.MemoryLimit
	} else {
		if cpuReq == "" {
			cpuReq = opts.CpuRequest
		}
		if memReq == "" {
			memReq = opts.MemoryRequest
		}
		if cpuLim == "" {
			cpuLim = opts.CpuLimit
		}
		if memLim == "" {
			memLim = opts.MemoryLimit
		}
	}

	var volumeMounts []VolumeMount
	if vmList, ok := oldData["volumeMounts"].([]any); ok && len(vmList) > 0 {
		for _, vm := range vmList {
			if vmMap, ok := vm.(map[string]any); ok {
				vmName, _ := vmMap["name"].(string)
				mountPath, _ := vmMap["mountPath"].(string)
				if vmName != "" && mountPath != "" {
					volumeMounts = append(volumeMounts, VolumeMount{
						Name:      vmName,
						MountPath: mountPath,
					})
				}
			}
		}
	}

	var volumes []Volume
	if vList, ok := oldData["volumes"].([]any); ok && len(vList) > 0 {
		for _, v := range vList {
			if vMap, ok := v.(map[string]any); ok {
				var vol Volume
				if name, ok := vMap["name"].(string); ok {
					vol.Name = name
				}
				if hpMap, ok := vMap["hostPath"].(map[string]any); ok {
					path, _ := hpMap["path"].(string)
					hpType, _ := hpMap["type"].(string)
					vol.HostPath = &HostPath{Path: path, Type: hpType}
				}
				if cmMap, ok := vMap["configMap"].(map[string]any); ok {
					vol.ConfigMap = cmMap
				}
				if secMap, ok := vMap["secret"].(map[string]any); ok {
					vol.Secret = secMap
				}
				if nfsMap, ok := vMap["nfs"].(map[string]any); ok {
					vol.NFS = nfsMap
				}
				if vol.Name != "" {
					volumes = append(volumes, vol)
				}
			}
		}
	}

	// Omit default timezone setup as they are already defined in values.yaml
	isDefaultVolumeMounts := func(vm []VolumeMount) bool {
		if len(vm) != 1 {
			return false
		}
		return vm[0].Name == "tz-config" && vm[0].MountPath == "/etc/localtime"
	}
	isDefaultVolumes := func(v []Volume) bool {
		if len(v) != 1 {
			return false
		}
		return v[0].Name == "tz-config" && v[0].HostPath != nil && v[0].HostPath.Path == "/usr/share/zoneinfo/America/La_Paz"
	}

	if isDefaultVolumeMounts(volumeMounts) || len(volumeMounts) == 0 {
		volumeMounts = nil
	}
	if isDefaultVolumes(volumes) || len(volumes) == 0 {
		volumes = nil
	}

	var securityContext map[string]any
	if sc, ok := oldData["securityContext"].(map[string]any); ok {
		securityContext = sc
	}

	deploy := Deploy{
		ImagePullSecrets:     imagePullSecrets,
		ReplicaCount:         replicaCount,
		Image:                deployImage,
		EnvVars:              envVars,
		SecurityContext:      securityContext,
		RevisionHistoryLimit: revisionHistoryLimit,
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
		VolumeMounts: volumeMounts,
		Volumes:      volumes,
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
	var ingressEntries []IngressEntry
	hasLegacyIngress := false
	if _, ok := oldData["ingresses"]; ok {
		hasLegacyIngress = true
	} else if _, ok := oldData["ingress"]; ok {
		hasLegacyIngress = true
	}

	if hasLegacyIngress {
		parseHosts := func(hostsList []any) []IngressHost {
			var hosts []IngressHost
			for _, hItem := range hostsList {
				hMap, ok := hItem.(map[string]any)
				if !ok {
					continue
				}
				hostStr, _ := hMap["host"].(string)

				if hostStr == "" || strings.Contains(hostStr, "<name>") {
					envShort := "dev"
					if strings.Contains(environment, "stg") || strings.Contains(environment, "staging") {
						envShort = "stg"
					} else if strings.Contains(environment, "release") || strings.Contains(environment, "prep") {
						envShort = "prep"
					} else if strings.Contains(environment, "prod") || strings.Contains(environment, "master") {
						envShort = "prod"
					}
					if envShort == "prod" {
						hostStr = "api.bg.com.bo"
					} else {
						hostStr = fmt.Sprintf("api.%s.bg.com.bo", envShort)
					}
				}

				var paths []IngressPath
				if pathsList, ok := hMap["paths"].([]any); ok {
					for _, pItem := range pathsList {
						if pMap, ok := pItem.(map[string]any); ok {
							pStr, _ := pMap["path"].(string)
							pType, _ := pMap["pathType"].(string)
							if pType == "" {
								pType = "ImplementationSpecific"
							}
							if pStr != "" {
								paths = append(paths, IngressPath{Path: pStr, PathType: pType})
							}
						} else if pStr, ok := pItem.(string); ok && pStr != "" {
							paths = append(paths, IngressPath{Path: pStr, PathType: "ImplementationSpecific"})
						}
					}
				}
				if len(paths) == 0 {
					paths = []IngressPath{
						{Path: "/" + nameOverride + "/", PathType: "ImplementationSpecific"},
					}
				}
				hosts = append(hosts, IngressHost{
					Host:  hostStr,
					Paths: paths,
				})
			}
			return hosts
		}

		if ingressesList, ok := oldData["ingresses"].([]any); ok && len(ingressesList) > 0 {
			for _, ingItem := range ingressesList {
				ingMap, ok := ingItem.(map[string]any)
				if !ok {
					continue
				}
				name, _ := ingMap["name"].(string)
				if name == "" {
					name = "default"
				}
				enabled := true
				if e, ok := ingMap["enabled"].(bool); ok {
					enabled = e
				}
				var annotations map[string]string
				if ann, ok := ingMap["annotations"].(map[string]any); ok {
					annotations = make(map[string]string)
					for k, v := range ann {
						if vStr, ok := v.(string); ok {
							annotations[k] = vStr
						}
					}
				}
				ingressClassName, _ := ingMap["ingressClassName"].(string)
				if ingressClassName == "" {
					if annotations != nil {
						if ic, ok := annotations["kubernetes.io/ingress.class"]; ok {
							ingressClassName = ic
						}
					}
				}
				if ingressClassName == "" {
					ingressClassName = "kong"
				}

				var hosts []IngressHost
				if hostsList, ok := ingMap["hosts"].([]any); ok {
					hosts = parseHosts(hostsList)
				}

				ingressEntries = append(ingressEntries, IngressEntry{
					Name:             name,
					Enabled:          enabled,
					IngressClassName: ingressClassName,
					Annotations:      annotations,
					Hosts:            hosts,
				})
			}
		} else if ingressMap, ok := oldData["ingress"].(map[string]any); ok {
			enabled := true
			if e, ok := ingressMap["enabled"].(bool); ok {
				enabled = e
			}
			var annotations map[string]string
			if ann, ok := ingressMap["annotations"].(map[string]any); ok {
				annotations = make(map[string]string)
				for k, v := range ann {
					if vStr, ok := v.(string); ok {
						annotations[k] = vStr
					}
				}
			}
			ingressClassName, _ := ingressMap["ingressClassName"].(string)
			if ingressClassName == "" {
				if annotations != nil {
					if ic, ok := annotations["kubernetes.io/ingress.class"]; ok {
						ingressClassName = ic
					}
				}
			}
			if ingressClassName == "" {
				ingressClassName = "kong"
			}

			var hosts []IngressHost
			if hostsList, ok := ingressMap["hosts"].([]any); ok {
				hosts = parseHosts(hostsList)
			} else {
				envShort := "dev"
				if strings.Contains(environment, "stg") || strings.Contains(environment, "staging") {
					envShort = "stg"
				} else if strings.Contains(environment, "release") || strings.Contains(environment, "prep") {
					envShort = "prep"
				} else if strings.Contains(environment, "prod") || strings.Contains(environment, "master") {
					envShort = "prod"
				}
				fallbackHost := ""
				if envShort == "prod" {
					fallbackHost = "api.bg.com.bo"
				} else {
					fallbackHost = fmt.Sprintf("api.%s.bg.com.bo", envShort)
				}
				hosts = []IngressHost{
					{
						Host: fallbackHost,
						Paths: []IngressPath{
							{Path: "/" + nameOverride + "/", PathType: "ImplementationSpecific"},
						},
					},
				}
			}

			ingressEntries = append(ingressEntries, IngressEntry{
				Name:             "default",
				Enabled:          enabled,
				IngressClassName: ingressClassName,
				Annotations:      annotations,
				Hosts:            hosts,
			})
		}
	}

	var kong *Kong
	if kongMap, ok := oldData["kong"].(map[string]any); ok {
		k := Kong{}
		if pluginsList, ok := kongMap["plugins"].([]any); ok && len(pluginsList) > 0 {
			for _, pItem := range pluginsList {
				if pMap, ok := pItem.(map[string]any); ok {
					var plugin KongPlugin
					plugin.Name, _ = pMap["name"].(string)
					plugin.Plugin, _ = pMap["plugin"].(string)
					if cfg, ok := pMap["config"].(map[string]any); ok {
						plugin.Config = cfg
					}
					k.Plugins = append(k.Plugins, plugin)
				}
			}
		}
		enabledVal := true
		if e, ok := kongMap["enabled"].(bool); ok {
			enabledVal = e
		}
		k.Enabled = &enabledVal
		kong = &k
	}

	networking := Networking{
		Ingress: ingressEntries,
		Kong:    kong,
	}

	webappValues := WebappValues{
		NameOverride:   nameOverride,
		Cluster:        opts.Cluster,
		Environment:    environment,
		Team:           team,
		Project:        project,
		ExtEnvVarsFrom: extEnvVarsFrom,
		Autoscaling:    autoscaling,
		Storage:        storage,
		Deploy:         deploy,
		Networking:     networking,
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(webappValues); err != nil {
		return "", "", fmt.Errorf("error serializing translated YAML: %w", err)
	}

	yamlHeader := "########################################\n## CONFIG | Web server values\n########################################\n"
	finalYaml := yamlHeader + unwrapYaml(buf.String())

	targetPath := fmt.Sprintf("charts/webapp/values/%s/%s/%s/%s.yaml", opts.Cluster, environment, team, nameOverride)

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
	return TranslateValuesWithArgoWithOptions(oldYamlStr, TranslationOptions{
		Cluster:             cluster,
		EnvOverride:         envOverride,
		TeamOverride:        teamOverride,
		JavaOptionsOverride: javaOptionsOverride,
		UseCommonConfigmap:  useCommonConfigmap,
	})
}

func TranslateValuesWithArgoWithOptions(oldYamlStr string, opts TranslationOptions) (string, string, string, string, string, error) {
	translatedValues, targetValuesPath, err := TranslateValuesWithOptions(oldYamlStr, opts)
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
	if opts.EnvOverride != "" {
		environment = opts.EnvOverride
	} else if v, ok := oldData["environment"].(string); ok && v != "" {
		environment = v
	}

	team := "middleware"
	if opts.TeamOverride != "" {
		team = opts.TeamOverride
	} else if v, ok := oldData["team"].(string); ok && v != "" {
		team = v
	}

	argoApp, argoPath, err := TranslateArgoAppFromParams(nameOverride, team, environment, opts.Cluster, true)
	if err != nil {
		return "", "", "", "", "", err
	}

	legacyArgoPath := fmt.Sprintf("argocd/%s/%s/%s/%s.yaml", opts.Cluster, environment, team, nameOverride)

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
