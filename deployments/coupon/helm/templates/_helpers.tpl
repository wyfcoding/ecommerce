{/*
Expand the name of the chart.
*/}
{- define "coupon.name" -}
{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }
{- end }

{/*
Create a default fully qualified app name.
*/}
{- define "coupon.fullname" -}
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
{- define "coupon.chart" -}
{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }
{- end }

{/*
Common labels
*/}
{- define "coupon.labels" -}
helm.sh/chart: { include "coupon.chart" . }
{ include "coupon.selectorLabels" . }
{- if .Chart.AppVersion }
app.kubernetes.io/version: { .Chart.AppVersion | quote }
{- end }
app.kubernetes.io/managed-by: { .Release.Service }
{- end }

{/*
Selector labels
*/}
{- define "coupon.selectorLabels" -}
app.kubernetes.io/name: { include "coupon.name" . }
app.kubernetes.io/instance: { .Release.Name }
{- end }

{/*
Create the name of the service account to use
*/}
{- define "coupon.serviceAccountName" -}
{- if .Values.serviceAccount.create }
{- default (include "coupon.fullname" .) .Values.serviceAccount.name }
{- else }
{- default "default" .Values.serviceAccount.name }
{- end }
{- end }
