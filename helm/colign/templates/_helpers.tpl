{{/*
Common labels
*/}}
{{- define "colign.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "colign.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database URL — external DB takes priority, falls back to Bitnami subchart
*/}}
{{- define "colign.databaseURL" -}}
{{- if .Values.externalDatabase.url -}}
  {{- .Values.externalDatabase.url -}}
{{- else if .Values.externalDatabase.host -}}
  postgres://{{ .Values.externalDatabase.username }}:{{ .Values.externalDatabase.password }}@{{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port | default 5432 }}/{{ .Values.externalDatabase.database | default "colign" }}?sslmode={{ .Values.externalDatabase.sslmode | default "disable" }}&search_path={{ .Values.externalDatabase.schema | default "public" }}
{{- else if .Values.postgresql.enabled -}}
  postgres://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ .Release.Name }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=disable
{{- end -}}
{{- end }}

{{/*
Redis URL — external Redis takes priority, falls back to Bitnami subchart
*/}}
{{- define "colign.redisURL" -}}
{{- if .Values.externalRedis.url -}}
  {{- .Values.externalRedis.url -}}
{{- else if .Values.externalRedis.host -}}
  {{- if .Values.externalRedis.password -}}
  redis://:{{ .Values.externalRedis.password }}@{{ .Values.externalRedis.host }}:{{ .Values.externalRedis.port | default 6379 }}
  {{- else -}}
  redis://{{ .Values.externalRedis.host }}:{{ .Values.externalRedis.port | default 6379 }}
  {{- end -}}
{{- else if .Values.redis.enabled -}}
  redis://{{ .Release.Name }}-redis-master:6379
{{- end -}}
{{- end }}
