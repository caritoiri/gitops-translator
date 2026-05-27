package translator

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const sampleInput = `
nameOverride: account-manager
environment: develop
team: middleware
replicaCount: 0
image:
  repository: bancoganadero/account-manager
  tag: 25080.develop
configmap: 
  spring.profiles.active: docker
`

const sampleSealed = `
nameOverride: account-manager
environment: develop
team: middleware
replicaCount: 0
image:
  repository: bancoganadero/account-manager
  tag: 25080.develop
sealed:
  data:
    SECRET_KEY: AgAWyABA...
`

func TestTranslateUseCommonConfigmap(t *testing.T) {
	output, path, err := TranslateValues(sampleInput, "on-premise", "", "", "-Xms256m", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]any
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	extEnv, ok := data["extEnvVarsFrom"].(map[string]any)
	if !ok {
		t.Fatalf("extEnvVarsFrom missing or invalid")
	}

	if extEnv["enabled"] != true {
		t.Errorf("expected enabled to be true, got %v", extEnv["enabled"])
	}

	envFrom, ok := extEnv["envFrom"].([]any)
	if !ok || len(envFrom) == 0 {
		t.Fatalf("envFrom missing or empty")
	}

	entry := envFrom[0].(map[string]any)
	cmRef := entry["configMapRef"].(map[string]any)
	if cmRef["name"] != "commun" {
		t.Errorf("expected configmap name to be commun, got %v", cmRef["name"])
	}

	if !strings.Contains(path, "on-premise/develop/middleware/account-manager.yaml") {
		t.Errorf("unexpected target path: %s", path)
	}
}

func TestTranslateNoCommonConfigmap(t *testing.T) {
	output, _, err := TranslateValues(sampleInput, "on-premise", "", "", "-Xms256m", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]any
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	extEnv, ok := data["extEnvVarsFrom"].(map[string]any)
	if !ok {
		t.Fatalf("extEnvVarsFrom missing or invalid")
	}

	if extEnv["enabled"] != false {
		t.Errorf("expected enabled to be false, got %v", extEnv["enabled"])
	}

	if _, exists := extEnv["envFrom"]; exists {
		t.Errorf("envFrom should not be present")
	}
}

const sampleSealedFull = `nameOverride: account-manager
environment: release
team: middleware
replicaCount: 1
image:
  pullPolicy: IfNotPresent
  repository: bancoganadero/account-manager
  tag: 24305.release
imagePullSecrets:
  - name: dockerhublogin
configmap:
  spring.profiles.active: docker
  java.awt.headless: false
  ORACLE_MAX_POOL: 2
  ORACLE_CONNECTION_TIMEOUT: 30000
  ORACLE_IDLE_TIMEOUT: 180000
  ORACLE_MINIMUM_IDLE: 2
  ORACLE_POOL_NAME: HIKARI-POOL
  PROJECT_SQL_PORT: 3306
  PROJECT_TIME_ZONE: America/La_Paz
  REDIS_PORT: 6379
  PROJECT_CACHE_TYPE: redis
  CONFIG_DATABASE_ENABLED: false
  CONFIG_SERVER_ENABLED: true
  CONFIG_SERVER_ADDR: http://configuration-manager:8080/configuration-manager
sealed:
  data: 
    ORACLE_DATABASE: AgAWyABA/emrNAEoz7WJzO9ExdeJSZs8XM0YYKX3QzspSaawbBmIkv0bR4WTLDuUCKPlBZLGYcjC5rp+5Gy79X5aDq7SofMW1++/4RhFGptkebI+5FIpYGj3wRXYZqHfupxkUeq3FGVeaQIw5ryanAIMtADvk2WPykD2p+HBx08gSdAr8+A5j78KI8tJaT8N7I99+sODIu2bi2m9UJMNCHBCB2ai90uneddkxPDO0tcPYjV0oIrnpaqRk+WHj3vk+u+ieeoujKvEBrCGzkclrsI4Zv6v8NlmAs2MGScM35qesb++dsmTmhThu0EXwjpYGjqxiAebfR7c66c1TP7qgIwdgdv8YufGHmlBY/OMet/F1ebv8EMtwsaiucxPy14rZqj97kLJ67p6vdVM+PLdE8LwtCB0qFSeFjpONCgbwWt4z64C/Jlh9UMw/JMxMI8VKhzOKjIPD2rKmAFtfWu4lg2QrWb28VA3aafFrgj/k7zJ4P+y0LwqIeqayR2B38xbaxBRXbHmAwsq1priuTu31LDCitWYL1jzOTRWXsjWxMBX/abybfwcVM7lcpnXsvMT2zvvmbWIlXi0hEV/52fZo1lkI9U7AHQJ96YRxd1a3lpw+6d+DTPhkfzlXJu2j+Px7Rjk/yft/jipqv/NflxJWQ4JMviE4uyKszM8BlhH7CZRbdww26AMtUSml1reikZ1cq37sjiVI/gxxrZH
    ORACLE_PASSWORD: AgC1Cqzrh69CSfXSsGsJHUE7VZQwfv5DGv5AFmMZPztvSXtynUD6pooz6fRDVCk5EbrtkawxEHJHOMSpYqa5lwaj6ovvl5oPMznsM7gWcfm/G0Wa25F19Phg2U+U1+kCkECwpyV8eUpZn77O0SMtk5TNO8rmzTTSFe7r6kA7WD07PV/+L7kLFpEfIJ8qehl/qk60eYlzg/hDFh7dk+fmS/j5OXMN/dwAH7Zp7OgdX3c5qdJUk/6STs7wkvqSybVHgnxaxN6AhCYT9morLW29jTbGlQd41mcHT/rEKXM/Ku0HzA+95Lp1mzpB9XeVkbOo11/lGntZRNubjBBFjtq8ThjwQIx1cYMKKheGyKvZf/E9bP8rYENMFybYoEc5lJpe1Md3j440DWHrWJRUjuCBR8qmVwcYRJaED3Q9WdY5TwFfb7kgDrST+ZpVcDv9I2EgUHUXueW2CaGq2UIqNYYfeCnzJp2rk1JVBaD4QCnJDxSlm1+yNwQ1p2wk/izpQSgO2Lyl5APSexcE2KeDcdvDOJweE3WuWAXrQTb5f1buHhXVKHiEnDTnpSb7gFsXBQsh2cdGYSI8B5kza33BtjKtoqCKJBUp4SWYp1tWkMKrCMaq8lmCmOS73T7wuTfgJkR50Kqsy+JbGN+nO9XxlyTADfn1dhLQMILok5pizUyzoid5gSSBCZwlQ+nHMiPSnqRN1nN1VCZv3/g=
    ORACLE_PORT: AgBLjXJr5T8mariSjDy6wKLDyJIh9we4Jc9Uzenj1rgQNzI6+HZpgi01S0XraAeSlIrJ7lu4iZhGM5eTD3AwrN9byLFp4+oTsBNdvyGXcUwWvgCeJj7Dv82e2Q1RqU36dvBLIQ0AmyUFfw++9LBx6btVZRfRmudV5kLZoxLqT4orAeMjufrHRyfAgPbrQRJsLllAzFi68xX+Lze6qQPsquJ9o/vu8ro3es9vyIkSieX+maVa3oWnT+sF4t+KmqASxTH9gxiO/z950Vq4Y6/irSwyjiIhxrOFEfMaQZVzfdW32uigwA6QhVWFRqFZJ8RIFGhEKpuWwAsH9WA5FrzSZJMmvLYcJuDpf6YVc2evlJQdYNYxcHTofZXH4ZWE+CGJAoH2UhIImDMDwgomV+FUQgBro8/jAYDypZJQVPaj+FOaptYtyV24vn613zbCRDcndrJvco4Dfuhb4kHK++ZHUqhtULbq9qZK2/Ytlvspo3RQKa0gwXNBOYk8vbmeQFm6RysSadTdIrkmYq96/J7mKjFf8599zIbFbdhUXLE0hYg5ynXwI30F2zAhst2Btk2AQsiD25l43k3KuAGtjkouSfth4ojJ6oqRhkupIAj8+4k+hX/e4xVqEi08T/iknQPYFf8iAKZkUC1r9x6/fudk6+0L6rA4eSKFnGUrsaHJieL4KQfje7/BopVAQCDlG1XyPtZj2vyo
    ORACLE_SERVER: AgCjnhcZksAxSaMUSEVZFtRS7vbGn642ddYsf2EtkifjqWCs6WDS3ELcmhJxRr/Z9ulEJNHAu2LafX26T/e4iLaNid333jleJdPCgD7RuPgtbBCUbbEA5fZDlg+rRmd9cI8R1bTHA9LzNdY2ZmpynAknqYzyxDvLP3g1l53bbTM1DqqOBR0BOcDhMKQbsYWPzA3NzozU8DWbBmNZ6AJh//FhG1vArMkWT7egA75ltyO04YM/OlvaBbjlFe+FhJPIzU6mMmgSbCUIKIjCg5b76v76h0gEUQDnEhG0rq093fG74KkCi+gapKr3E56udzHju5wobrpRxRsJgdquPov7Rcd3pDEZat+v8OuZbLiRsvaSTQGT8ylozIgUcPSBXb1mxc9Fio2RNHBh1xUa9jE06/xETFQKzOB5mbxFWX/eBf0gdWXjksLE1Y2x2eGA4YO+2VQAeToMQsevE16+djEi5I18ucZldqo4jbS1Pn1PyPeVOcxv1fwFky+JwlyK3h4mz1jCDmUgV2wYb1qSARzvBONOt+A97q5Rrt93Xs5oxLkxxEOxf2zHKRzXk/3JHewJElroIJgOUWy8Eosq+o4CrkLho3KNpUbQq26BiptC+q1SDL/p80HWWbAbaPy5dpxK8CcKodnPsDZd8xx4VsVk2GW+RXbrWc8S30AgRZcWOESbmr6OgRpR0W7HQV38x5pmOW1JZ3zr5IKaOfr+5OpY6R66gw==
    ORACLE_USER: AgArUWICFh4zdbFW5HVyVt2k1WTYFi655hLNCZHVjb1vVMpREoLX14aiNT+TXVM8InhHgigwU4UtwxQHqFQJPjF4PF/9sfxZwov1oF4lxS1IByjr2kSJmNCR4yaKKo01FbZyqPCyMF3I6bWm8+6ZdQmfCb4peujr9OsQNTMsQundsUGx3k6penI0Lu0n1FVPfTdIsamEMqMfAAgVrATRuH3OwKpBvVKdAvvtHu3RoHVE9CD7UbNqFRI4wVdVBFzszbp4y+d7gHCZ35harntHgOO0KUBqF3yd81lJeW0kPHzgt3aJ+KpARm9c+lVECfhnA9jpFw73MhUORdQrVMJCn3O7c1v69vXnpV9TYjCwi4NU66l7oYgkO0y6nqYpu5lIQigcIHKa+MU9Um7UOuYFRNHIx0Sq4UlOyXi+OyoDAcjNIMQQGIE0fqpTErOF59XSLuZdp+mintz+ai4bHD/+ALIW4jQEfFuNqdhKmkO1v+z03dYJrWEQ4IE82QSJt7tJDTX+60jPo8+3lfyuvXRFc81gf7Sv5LgO01kLll1wB+Oc6i4K6d0aNm2C0es7hASnIPhIOUokU7d1Jv3B6sq7OdgcUpS1pjGFA/9tSqaHDnsXA3pzIx9CaXdz4gqA8aCZuhbh2IXL4QjBrryg8iMK2TFsjf604g4/hLoGj3ZqYwNcI4JLlm5inMMSJXCcFERntl6jsq+DMahj9
    REDIS_SERVER: AgAeIvNtGKXYPc90GtYwLV6MTpHw8zot0XvomOlvr9CgrFNaKgAbqcC2LqstEyDq35qTLpL/UJl2BLYZCgzGny6G6S6H1sD65HyUZiPLldqyFmnCqfDsCONZdkT7L+zleaFonF/4BYB37ueK9GAPVTTXiTia6A9vO0tXFE30jAHcnGvfNhQ61wwkAp1ca8dFpfD/Q9rP5hh0sRnxjAsFQhtVByUJndY8gfjrs6dEf0CopkQ7gxjLxrTv635Nf2VKCsRWpDMtLY5p/Uwg4ftvKiloKh0UbKJIlulBz+aF2G//4YJwuiQDvCLzin29HFsQxnkL5ovtApwbKXjKzDreIEVqW+gkvOfYpVJQdnUSQ9RAxXl0Lgb0oN8kZQj1OWUOIMbMS0rtrZA7B7pJTM01F8fp1LoF3lu5X5UYm6BuvAHknfaDwLEMkx+VxlRr1TXlp96hdDiL0aQc9hcFakPd+KOMJ73BflUm3KHlaRhjo4du/PaZ5KSdD71KOe1XyfEjbyf7g3dNwLu+KnqHXoJ2vR/1yZauzuaw0bUMM7RuW62JmAvCeYB7DVYX3KGHfsZDWF3NRasZL0hoyb8bHcw/KzWz2yQBXKCaiVEUhs1Iki8PBx5EXhYZDulGFwV3SZ/mgjfTf8KXX3+Inkqi+ygzpjBjc9HaYHDHmS14pYl4wRWZUsjplWNmnhSkBNRLdsBlQPLZlwlGpCyl1VCTx5+4L2bTO78E4OWxkMPrNm0EQ9onfSA+elE=
ingress:
  enabled: false
  annotations:
    kubernetes.io/ingress.class: kong
    konghq.com/strip-path: "true"
  hosts:
    - paths:
        - path: /middleware/account-manager/
          pathType: Prefix
`

func TestTranslateFullSealedSecrets(t *testing.T) {
	output, _, err := TranslateValues(sampleSealedFull, "on-premise", "", "", "-Xms256m", false)
	if err != nil {
		t.Fatalf("unexpected error translating full sealed secrets: %v", err)
	}

	t.Logf("Translated output:\n%s", output)

	var data map[string]any
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	extEnv, ok := data["extEnvVarsFrom"].(map[string]any)
	if !ok {
		t.Fatalf("extEnvVarsFrom missing or invalid")
	}

	if extEnv["enabled"] != true {
		t.Errorf("expected enabled to be true, got %v", extEnv["enabled"])
	}
}

func TestUnwrapYaml(t *testing.T) {
	input := `secretRef:
    DATABASE: firstLine
continuationLine
    PASSWORD: secondLine`
	expected := `secretRef:
    DATABASE: firstLinecontinuationLine
    PASSWORD: secondLine`
	output := unwrapYaml(input)
	if output != expected {
		t.Errorf("unwrapYaml failed.\nGot:\n%s\nExpected:\n%s", output, expected)
	}
}

const sampleArgoLegacy = `# ---
# apiVersion: argoproj.io/v1alpha1
# kind: Application
# metadata:
#   name: bcb-spt-connector
#   namespace: argocd
# spec:
#   project: az-devops-gitops
#   source:
#     repoURL: https://BancoGanadero@dev.azure.com/BancoGanadero/BGA-Internal/_git/gitops
#     path: infra/bcb/spt-connector
#     targetRevision: main
#     helm:
#       valueFiles:
#         - values-develop.yaml
#   destination:
#     server: https://kubernetes.default.svc
#     namespace: bcb`

func TestTranslateArgoAppCommented(t *testing.T) {
	output, path, delPath, err := TranslateArgoApp(sampleArgoLegacy, "on-premise", true)
	if err != nil {
		t.Fatalf("unexpected error translating legacy Argo app: %v", err)
	}

	if !strings.Contains(path, "cluster/on-premise/develop/apps/bcb/spt-connector.yaml") {
		t.Errorf("expected target path to be cluster/on-premise/develop/apps/bcb/spt-connector.yaml, got %s", path)
	}

	if delPath != "argocd/on-premise/develop/bcb/spt-connector.yaml" {
		t.Errorf("expected deletion path to be argocd/on-premise/develop/bcb/spt-connector.yaml, got %s", delPath)
	}

	if !strings.Contains(output, "name: bcb-spt-connector") {
		t.Errorf("expected translated app name to be bcb-spt-connector, got:\n%s", output)
	}

	if !strings.Contains(output, "project: bcb") {
		t.Errorf("expected translated project to be bcb, got:\n%s", output)
	}

	if !strings.Contains(output, "syncPolicy:") {
		t.Errorf("expected translated app to have syncPolicy, got:\n%s", output)
	}
}

func TestTranslateArgoAppCommentedOutSyncPolicy(t *testing.T) {
	output, _, _, err := TranslateArgoApp(sampleArgoLegacy, "on-premise", false)
	if err != nil {
		t.Fatalf("unexpected error translating legacy Argo app: %v", err)
	}

	if !strings.Contains(output, "# syncPolicy:") {
		t.Errorf("expected translated app to have commented out syncPolicy, got:\n%s", output)
	}
}

