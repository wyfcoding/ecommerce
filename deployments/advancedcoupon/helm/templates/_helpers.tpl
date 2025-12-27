{/*
Expand the name of the chart.
*/}
{- define "advancedcoupon.name" -}
{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }
{- end }

{/*
Create a default fully qualified app name.
*/}
{- define "advancedcoupon.fullname" -}
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
{- define "advancedcoupon.chart" -}
{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }
{- end }

{/*
Common labels
*/}
{- define "advancedcoupon.labels" -}
helm.sh/chart: { include "advancedcoupon.chart" . }
{ include "advancedcoupon.selectorLabels" . }
{- if .Chart.AppVersion }
app.kubernetes.io/version: { .Chart.AppVersion | quote }
{- end }
app.kubernetes.io/managed-by: { .Release.Service }
{- end }

{/*
Selector labels
*/}
{- define "advancedcoupon.selectorLabels" -}
app.kubernetes.io/name: { include "advancedcoupon.name" . }
app.kubernetes.io/instance: { .Release.Name }
{- end }

{/*
Create the name of the service account to use
*/}
{- define "advancedcoupon.serviceAccountName" -}
{- if .Values.serviceAccount.create }
{- default (include "advancedcoupon.fullname" .) .Values.serviceAccount.name }
{- else }
{- default "default" .Values.serviceAccount.name }
{- end }
{- end }
