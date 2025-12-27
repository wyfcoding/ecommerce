{/*
Expand the name of the chart.
*/}
{- define "dataprocessing.name" -}
{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }
{- end }

{/*
Create a default fully qualified app name.
*/}
{- define "dataprocessing.fullname" -}
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
{- define "dataprocessing.chart" -}
{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }
{- end }

{/*
Common labels
*/}
{- define "dataprocessing.labels" -}
helm.sh/chart: { include "dataprocessing.chart" . }
{ include "dataprocessing.selectorLabels" . }
{- if .Chart.AppVersion }
app.kubernetes.io/version: { .Chart.AppVersion | quote }
{- end }
app.kubernetes.io/managed-by: { .Release.Service }
{- end }

{/*
Selector labels
*/}
{- define "dataprocessing.selectorLabels" -}
app.kubernetes.io/name: { include "dataprocessing.name" . }
app.kubernetes.io/instance: { .Release.Name }
{- end }

{/*
Create the name of the service account to use
*/}
{- define "dataprocessing.serviceAccountName" -}
{- if .Values.serviceAccount.create }
{- default (include "dataprocessing.fullname" .) .Values.serviceAccount.name }
{- else }
{- default "default" .Values.serviceAccount.name }
{- end }
{- end }
