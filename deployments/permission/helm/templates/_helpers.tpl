{/*
Expand the name of the chart.
*/}
{- define "permission.name" -}
{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }
{- end }

{/*
Create a default fully qualified app name.
*/}
{- define "permission.fullname" -}
{- if .Values.fullnameOverride }
{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }
{- else }
{- $name := default .Chart.Name .Values.nameOverride }
{- if contains $name .Release.Name }
{- .Release.Name | trunc 63 | trimSuffix "-" }
{- else }
{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }
{- end }
{- end }
{- end }

{/*
Create chart name and version as used by the chart label.
*/}
{- define "permission.chart" -}
{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }
{- end }

{/*
Common labels
*/}
{- define "permission.labels" -}
helm.sh/chart: { include "permission.chart" . }
{ include "permission.selectorLabels" . }
{- if .Chart.AppVersion }
app.kubernetes.io/version: { .Chart.AppVersion | quote }
{- end }
app.kubernetes.io/managed-by: { .Release.Service }
{- end }

{/*
Selector labels
*/}
{- define "permission.selectorLabels" -}
app.kubernetes.io/name: { include "permission.name" . }
app.kubernetes.io/instance: { .Release.Name }
{- end }

{/*
Create the name of the service account to use
*/}
{- define "permission.serviceAccountName" -}
{- if .Values.serviceAccount.create }
{- default (include "permission.fullname" .) .Values.serviceAccount.name }
{- else }
{- default "default" .Values.serviceAccount.name }
{- end }
{- end }
