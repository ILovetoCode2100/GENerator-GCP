{{/*
Expand the name of the chart.
*/}}
{{- define "virtuoso-api-cli.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "virtuoso-api-cli.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "virtuoso-api-cli.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "virtuoso-api-cli.labels" -}}
helm.sh/chart: {{ include "virtuoso-api-cli.chart" . }}
{{ include "virtuoso-api-cli.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "virtuoso-api-cli.selectorLabels" -}}
app.kubernetes.io/name: {{ include "virtuoso-api-cli.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "virtuoso-api-cli.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "virtuoso-api-cli.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret
*/}}
{{- define "virtuoso-api-cli.secretName" -}}
{{- if .Values.externalSecrets.enabled }}
{{- printf "%s-external" (include "virtuoso-api-cli.fullname" .) }}
{{- else if .Values.secret.create }}
{{- default (printf "%s-secret" (include "virtuoso-api-cli.fullname" .)) .Values.secret.name }}
{{- else }}
{{- default (printf "%s-secret" (include "virtuoso-api-cli.fullname" .)) .Values.secret.name }}
{{- end }}
{{- end }}

{{/*
Get the redis host
*/}}
{{- define "virtuoso-api-cli.redisHost" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master" (include "virtuoso-api-cli.fullname" .) }}
{{- else }}
{{- .Values.config.redis.host }}
{{- end }}
{{- end }}

{{/*
Get the redis port
*/}}
{{- define "virtuoso-api-cli.redisPort" -}}
{{- if .Values.redis.enabled }}
{{- .Values.redis.master.service.ports.redis | default 6379 }}
{{- else }}
{{- .Values.config.redis.port }}
{{- end }}
{{- end }}

{{/*
Get the image
*/}}
{{- define "virtuoso-api-cli.image" -}}
{{- $registryName := .Values.image.registry | default .Values.global.imageRegistry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "virtuoso-api-cli.imagePullSecrets" -}}
{{- $pullSecrets := list }}
{{- range .Values.global.imagePullSecrets -}}
  {{- $pullSecrets = append $pullSecrets . -}}
{{- end -}}
{{- range .Values.imagePullSecrets -}}
  {{- $pullSecrets = append $pullSecrets . -}}
{{- end -}}
{{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
{{- range $pullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Compile all warnings into a single message.
*/}}
{{- define "virtuoso-api-cli.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "virtuoso-api-cli.validateValues.apiToken" .) -}}
{{- $messages := append $messages (include "virtuoso-api-cli.validateValues.organization" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message -}}
{{- end -}}
{{- end -}}

{{/*
Validate values - API token
*/}}
{{- define "virtuoso-api-cli.validateValues.apiToken" -}}
{{- if and .Values.secret.create (not .Values.externalSecrets.enabled) -}}
{{- if not .Values.secret.apiToken -}}
virtuoso-api-cli: apiToken
    You must provide an API token when creating a secret.
    Please set the secret.apiToken value.
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Validate values - Organization ID
*/}}
{{- define "virtuoso-api-cli.validateValues.organization" -}}
{{- if not .Values.config.organization.id -}}
virtuoso-api-cli: organization.id
    You must provide an organization ID.
    Please set the config.organization.id value.
{{- end -}}
{{- end -}}
