{/*
Expand the name of the chart.
*/}
{- define "orderoptimization.name" -}
{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }
{- end }

{/*
Create a default fully qualified app name.
*/}
{- define "orderoptimization.fullname" -}
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
{- define "orderoptimization.chart" -}
{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }
{- end }

{/*
Common labels
*/}
{- define "orderoptimization.labels" -}
helm.sh/chart: { include "orderoptimization.chart" . }
{ include "orderoptimization.selectorLabels" . }
{- if .Chart.AppVersion }
app.kubernetes.io/version: { .Chart.AppVersion | quote }
{- end }
app.kubernetes.io/managed-by: { .Release.Service }
{- end }

{/*
Selector labels
*/}
{- define "orderoptimization.selectorLabels" -}
app.kubernetes.io/name: { include "orderoptimization.name" . }
app.kubernetes.io/instance: { .Release.Name }
{- end }

{/*
Create the name of the service account to use
*/}
{- define "orderoptimization.serviceAccountName" -}
{- if .Values.serviceAccount.create }
{- default (include "orderoptimization.fullname" .) .Values.serviceAccount.name }
{- else }
{- default "default" .Values.serviceAccount.name }
{- end }
{- end }
