package main

import (
	"fmt"
	"gitops-values-translator-go/translator" // Tu importación local corregida
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ResourceProfile struct {
	CpuRequest    string
	MemoryRequest string
	CpuLimit      string
	MemoryLimit   string
}

// Perfiles de recursos por tecnología para evitar configurar valores manuales repetitivos.
var ResourceProfiles = map[string]ResourceProfile{
	"java":    {"250m", "512Mi", "500m", "1Gi"},
	"nodejs":  {"100m", "256Mi", "300m", "512Mi"},
	"go":      {"50m", "64Mi", "200m", "256Mi"},
	"python":  {"100m", "128Mi", "400m", "512Mi"},
	"minimal": {"50m", "64Mi", "100m", "128Mi"},
}

type MigrationTarget struct {
	Name               string // Dejar en blanco, "*" o "all" para migrar todo el namespace
	Namespace          string
	Env                string
	UseCommonConfigmap bool
	JavaOptions        string // Opción para sobreescribir JAVA_TOOL_OPTIONS (vacío si no se requiere)
	Profile            string // "java", "nodejs", "go", "python", "minimal". Por defecto "java"
	ForceResources     bool   // Si es true, sobreescribe los recursos legacy con los definidos en el perfil o manuales
	Overwrite          bool   // Si es true, sobreescribe los archivos Helm Values y Argo App si ya existen

	// Sobreescrituras manuales específicas (tienen prioridad sobre el perfil si no están vacías)
	CpuRequest    string
	MemoryRequest string
	CpuLimit      string
	MemoryLimit   string
}

func main() {
	// Lista de servicios a migrar
	targets := []MigrationTarget{
		// ==================================================================================
		// EJEMPLOS DE CASOS DE USO (Descomentar y adaptar según sea necesario):
		// ==================================================================================

		// CASO DE USO 1: Migrar app individual SIN sobreescribir (Safe/Skip mode)
		// -> Si el archivo destino existe, se omite y no se realizan commits.
		// {"trxz-conector", "integration", "develop", false, "", "nodejs", false, false, "", "", "", ""},

		// CASO DE USO 2: Migrar app individual CON sobreescritura (Para corregir migraciones incorrectas)
		// -> Escribe los archivos siempre y los sube a Git.
		// {"trxz-conector", "integration", "develop", false, "", "nodejs", false, true, "", "", "", ""},

		// CASO DE USO 3: Migrar TODO un namespace usando el comodín "*" (ó "all" ó dejando vacío "")
		// -> Auto-detecta todos los archivos legacy del namespace en el ambiente indicado.
		// {"*", "integration", "develop", false, "", "nodejs", false, false, "", "", "", ""},

		// CASO DE USO 4: Migrar TODO un namespace con sobreescritura y perfiles de recursos
		// -> Sobreescribe todos los apps encontrados en el namespace con el perfil nodejs.
		// {"*", "integration", "develop", false, "", "nodejs", true, true, "", "", "", ""},

		// CASO DE USO 5: Migrar app Java forzando los recursos correctos de Java (250m-500m / 512Mi-1Gi)
		// -> Reemplaza los recursos legacy (si eran muy bajos) con los del perfil Java.
		// {"trxz-conector", "integration", "staging", false, "", "java", true, true, "", "", "", ""},

		// CASO DE USO 6: Migrar con recursos configurados de forma manual (Prioridad sobre perfil)
		// -> Especifica valores manuales de CPU y Memoria Request/Limit.
		// {"trxz-conector", "integration", "develop", false, "", "nodejs", true, true, "150m", "256Mi", "300m", "512Mi"},

		// ==================================================================================
		// EJECUCIÓN ACTIVA:
		// ==================================================================================
		{"api-bs-auth-login-web-app", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"api-bs-auth-login-web-app", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"api-bs-auth-login-web-app", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},

		{"bff-qrapido", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"bff-qrapido", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"bff-qrapido", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},

		{"core-remittance-manager", "rutinas-masivas", "develop", false, "", "nodejs", true, true, "", "", "", ""},
		{"core-remittance-manager", "rutinas-masivas", "staging", false, "", "nodejs", true, true, "", "", "", ""},
		{"core-remittance-manager", "rutinas-masivas", "release", false, "", "nodejs", true, true, "", "", "", ""},

		{"entel-adapter-conector", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"entel-adapter-conector", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"entel-adapter-conector", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},

		{"listas-negras-conector", "rutinas-masivas", "develop", false, "", "nodejs", true, true, "", "", "", ""},
		{"listas-negras-conector", "rutinas-masivas", "staging", false, "", "nodejs", true, true, "", "", "", ""},
		{"listas-negras-conector", "rutinas-masivas", "release", false, "", "nodejs", true, true, "", "", "", ""},

		{"migration-vpay-cre-manager", "rutinas-masivas", "develop", false, "", "nodejs", true, true, "", "", "", ""},
		{"migration-vpay-cre-manager", "rutinas-masivas", "staging", false, "", "nodejs", true, true, "", "", "", ""},
		{"migration-vpay-cre-manager", "rutinas-masivas", "release", false, "", "nodejs", true, true, "", "", "", ""},

		{"process-notification-connector", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"process-notification-connector", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"process-notification-connector", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},

		{"process-retries-worker", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"process-retries-worker", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"process-retries-worker", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},

		{"process-transaction-worker", "rutinas-masivas", "develop", false, "", "java", true, true, "", "", "", ""},
		{"process-transaction-worker", "rutinas-masivas", "staging", false, "", "java", true, true, "", "", "", ""},
		{"process-transaction-worker", "rutinas-masivas", "release", false, "", "java", true, true, "", "", "", ""},
	}

	cluster := "on-premise"

	// Lista final expandida de servicios a migrar
	var servicesToMigrate []MigrationTarget

	for _, t := range targets {
		if t.Name != "" && t.Name != "*" && strings.ToLower(t.Name) != "all" {
			servicesToMigrate = append(servicesToMigrate, t)
			continue
		}

		// Si el nombre está vacío, "*" o "all", buscamos todos los YAMLs en el namespace para migrar todo el directorio
		dirPath := fmt.Sprintf("../gitops/argocd/%s/%s/%s", cluster, t.Env, t.Namespace)
		files, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Printf("  [ERROR] No se pudo leer el directorio de Argo Apps %s para migración masiva: %v\n", dirPath, err)
			continue
		}

		fmt.Printf("  [INFO] Auto-detectando apps en namespace [%s] bajo ambiente [%s]...\n", t.Namespace, t.Env)
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
				appName := strings.TrimSuffix(f.Name(), ".yaml")
				servicesToMigrate = append(servicesToMigrate, MigrationTarget{
					Name:               appName,
					Namespace:          t.Namespace,
					Env:                t.Env,
					UseCommonConfigmap: t.UseCommonConfigmap,
					JavaOptions:        t.JavaOptions,
					Profile:            t.Profile,
					ForceResources:     t.ForceResources,
					Overwrite:          t.Overwrite,
					CpuRequest:         t.CpuRequest,
					MemoryRequest:      t.MemoryRequest,
					CpuLimit:           t.CpuLimit,
					MemoryLimit:        t.MemoryLimit,
				})
				fmt.Printf("    -> Detectada app para migración: %s\n", appName)
			}
		}
	}

	for _, svc := range servicesToMigrate {
		fmt.Printf("\n>>> Procesando migración de [%s] para ambiente [%s]...\n", svc.Name, svc.Env)

		// 1. Apuntar al repositorio GitOps hermano subiendo un nivel con "../"
		legacyPath := fmt.Sprintf("../gitops/argocd/%s/%s/%s/%s.yaml", cluster, svc.Env, svc.Namespace, svc.Name)
		legacyData, err := os.ReadFile(legacyPath)
		if err != nil {
			fmt.Printf("  [AVISO] No se encontró el archivo legacy en %s. Saltando...\n", legacyPath)
			continue
		}

		// Intentar parsear la Argo CD App legacy para encontrar la ruta real del archivo de Helm Values
		var legacyApp translator.LegacyApplication
		strippedYaml := stripComments(string(legacyData))
		var valuesPath string

		if err := yaml.Unmarshal([]byte(strippedYaml), &legacyApp); err == nil {
			path := legacyApp.Spec.Source.Path
			var valuesFile string
			if len(legacyApp.Spec.Source.Helm.ValueFiles) > 0 {
				valuesFile = legacyApp.Spec.Source.Helm.ValueFiles[0]
			} else {
				valuesFile = fmt.Sprintf("values-%s.yaml", svc.Env)
			}
			if path != "" && valuesFile != "" {
				valuesPath = filepath.Join("..", "gitops", path, valuesFile)
			}
		}

		// Fallback si no se pudo determinar por el parseo de la Argo App
		if valuesPath == "" {
			valuesPath = fmt.Sprintf("../gitops/infra/%s/%s/values-%s.yaml", svc.Namespace, svc.Name, svc.Env)
		}

		fmt.Printf("  [INFO] Leyendo valores legacy de: %s\n", valuesPath)
		valuesData, err := os.ReadFile(valuesPath)
		if err != nil {
			fmt.Printf("  [ERROR] No se pudo leer el archivo de valores legacy en %s: %v. Saltando...\n", valuesPath, err)
			continue
		}

		// Determinar perfil y valores de recursos
		profileName := strings.ToLower(strings.TrimSpace(svc.Profile))
		if profileName == "" {
			profileName = "java"
		}
		profile, exists := ResourceProfiles[profileName]
		if !exists {
			profile = ResourceProfiles["java"]
		}

		cpuReq := svc.CpuRequest
		if cpuReq == "" {
			cpuReq = profile.CpuRequest
		}
		memReq := svc.MemoryRequest
		if memReq == "" {
			memReq = profile.MemoryRequest
		}
		cpuLim := svc.CpuLimit
		if cpuLim == "" {
			cpuLim = profile.CpuLimit
		}
		memLim := svc.MemoryLimit
		if memLim == "" {
			memLim = profile.MemoryLimit
		}

		opts := translator.TranslationOptions{
			Cluster:             cluster,
			EnvOverride:         svc.Env,
			TeamOverride:        svc.Namespace,
			JavaOptionsOverride: svc.JavaOptions,
			UseCommonConfigmap:  svc.UseCommonConfigmap,
			CpuRequest:          cpuReq,
			MemoryRequest:       memReq,
			CpuLimit:            cpuLim,
			MemoryLimit:         memLim,
			ForceResourceLimits: svc.ForceResources,
		}

		// 2. Ejecutar el traductor pasándole los valores Helm legados y opciones
		translatedValues, targetValuesPath, argoApp, argoPath, _, err := translator.TranslateValuesWithArgoWithOptions(
			string(valuesData),
			opts,
		)
		if err != nil {
			fmt.Printf("  [ERROR] Al traducir el servicio %s: %v\n", svc.Name, err)
			continue
		}

		// --- CORRECCIÓN DE "unknown-app" EN CALIENTE ---
		// Si el traductor no encontró nameOverride y generó "unknown-app", lo sobreescribimos con el nombre correcto.
		if strings.Contains(targetValuesPath, "unknown-app") {
			fmt.Printf("  [AVISO] Detectado fallback 'unknown-app'. Corrigiendo rutas con el nombre real: '%s'\n", svc.Name)
			targetValuesPath = strings.ReplaceAll(targetValuesPath, "unknown-app", svc.Name)
			argoPath = strings.ReplaceAll(argoPath, "unknown-app", svc.Name)
			// Reemplazar la propiedad nameOverride dentro del contenido del YAML de Argo App y Helm Values
			translatedValues = strings.ReplaceAll(translatedValues, "nameOverride: unknown-app", "nameOverride: "+svc.Name)
			argoApp = strings.ReplaceAll(argoApp, "name: "+svc.Namespace+"-unknown-app", "name: "+svc.Namespace+"-"+svc.Name)
			argoApp = strings.ReplaceAll(argoApp, "unknown-app.yaml", svc.Name+".yaml")
		}

		// --- RESOLUCIÓN DE RUTAS PARA REPOSITORIOS VECINOS ---
		localTargetValuesPath := filepath.Join("..", targetValuesPath) // Ejemplo: ../charts/webapp/...
		localArgoPath := filepath.Join("../argocd", argoPath)          // Ejemplo: ../argocd/cluster/...

		hasChanges := false // Bandera para saber si realmente debemos hacer commit en Git

		// --- CONTROL INDEPENDIENTE PASO A PASO ---

		// Paso A: Escribir Helm Values en 'charts'
		if _, err := os.Stat(localTargetValuesPath); os.IsNotExist(err) || svc.Overwrite {
			err = os.MkdirAll(filepath.Dir(localTargetValuesPath), 0755)
			if err != nil {
				fmt.Printf("  [ERROR] Al crear directorios de destino %s: %v\n", localTargetValuesPath, err)
				continue
			}
			_ = os.WriteFile(localTargetValuesPath, []byte(translatedValues), 0644)
			fmt.Printf("  [OK] Creado/Sobreestrito Helm Values: %s\n", localTargetValuesPath)
			hasChanges = true
		} else {
			fmt.Printf("  [OMITIDO] El archivo de Helm Values ya existe: %s (No se sobreescribirá)\n", localTargetValuesPath)
		}

		// Paso B: Escribir Argo App en 'argocd'
		if _, err := os.Stat(localArgoPath); os.IsNotExist(err) || svc.Overwrite {
			err = os.MkdirAll(filepath.Dir(localArgoPath), 0755)
			if err != nil {
				fmt.Printf("  [ERROR] Al crear directorios de Argo App %s: %v\n", localArgoPath, err)
				continue
			}
			_ = os.WriteFile(localArgoPath, []byte(argoApp), 0644)
			fmt.Printf("  [OK] Creada/Sobreestrita Argo App: %s\n", localArgoPath)
			hasChanges = true
		} else {
			fmt.Printf("  [OMITIDO] La Argo App ya existe: %s (No se sobreescribirá)\n", localArgoPath)
		}

		// Paso C: Comentar el archivo legacy original en 'gitops' si aún no está comentado
		wasCommented, err := commentLegacyFile(legacyPath)
		if err != nil {
			fmt.Printf("  [ERROR] Al procesar el archivo legacy %s: %v\n", legacyPath, err)
		} else if wasCommented {
			fmt.Printf("  [OK] Comentado archivo legacy original: %s\n", legacyPath)
			hasChanges = true
		} else {
			fmt.Println("  [OMITIDO] El archivo legacy ya se encontraba comentado previamente.")
		}

		// Paso D: Ejecutar Git workflow únicamente si hubo cambios físicos reales
		if hasChanges {
			commitMessage := fmt.Sprintf("feat: migrated %s %s", svc.Name, svc.Env)
			fmt.Println("  [GIT] Enviando cambios detectados a los repositorios...")
			if err := runGitCommand("../gitops", commitMessage); err != nil {
				fmt.Printf("  [ERROR GIT] %v\n", err)
			}
			if err := runGitCommand("../argocd", commitMessage); err != nil {
				fmt.Printf("  [ERROR GIT] %v\n", err)
			}
			if err := runGitCommand("../charts", commitMessage); err != nil {
				fmt.Printf("  [ERROR GIT] %v\n", err)
			}
		} else {
			fmt.Println("  [GIT] Sin cambios pendientes. No se requiere ejecución de commits.")
		}
	}
}

// Retorna true si modificó el archivo, false si ya estaba comentado
func commentLegacyFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(data)
	// Si todas las líneas con contenido útil ya empiezan con '#', no hacemos nada
	lines := strings.Split(content, "\n")
	alreadyCommented := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "#") && trimmed != "---" {
			alreadyCommented = false
			break
		}
	}

	if alreadyCommented {
		return false, nil
	}

	// Comentar el archivo
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "#") && trimmed != "---" {
			lines[i] = "# " + line
		}
	}
	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	return true, err
}

func stripComments(yamlStr string) string {
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

func hasChangesToCommit(dir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func runGitCommand(dir string, message string) error {
	fmt.Printf("  [GIT] [%s] Agregando cambios: git add . ...\n", dir)
	if err := execCmd(dir, "git", "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	if !hasChangesToCommit(dir) {
		fmt.Printf("  [GIT] [%s] Sin cambios detectados para confirmar.\n", dir)
		return nil
	}

	fmt.Printf("  [GIT] [%s] Confirmando cambios: git commit -m \"%s\" ...\n", dir, message)
	if err := execCmd(dir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Printf("  [GIT] [%s] Sincronizando repositorio: git pull --rebase origin main...\n", dir)
	if err := execCmd(dir, "git", "pull", "--rebase", "origin", "main"); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	fmt.Printf("  [GIT] [%s] Enviando al servidor: git push origin main ...\n", dir)
	if err := execCmd(dir, "git", "push", "origin", "main"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	fmt.Printf("  [GIT] [%s] Repositorio actualizado con éxito.\n", dir)
	return nil
}

func execCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
