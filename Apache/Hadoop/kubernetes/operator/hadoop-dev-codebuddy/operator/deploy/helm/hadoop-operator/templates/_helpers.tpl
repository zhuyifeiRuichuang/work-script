{{/*
Expand the name of the chart.
*/}}
{{- $name := include "hadoop-operator.name" . -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- $fullname := include "hadoop-operator.fullname" . -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- $chart := printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}

{{/*
Common labels
*/}}
{{- define "hadoop-operator.labels" -}}
helm.sh/chart: {{ $chart }}
{{ include "hadoop-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hadoop-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ $name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Operator labels
*/}}
{{- define "hadoop-operator.operatorLabels" -}}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hadoop-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hadoop-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Define the fullname template
*/}}
{{- define "hadoop-operator.fullname" -}}
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
