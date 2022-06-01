# ------------------------------------------------------------------------------------------
### STEP 1
# ------------------------------------------------------------------------------------------

## Definer relevant namespace
NAMESPACE=monitoring

## Skaff IDen til relevant volum
PV_ID=$(kubectl -n ${NAMESPACE} get pvc | xargs | cut -d' ' -f 11)

## Skaff regionen (og AZ) volumet bor på
PV_REGION=$(kubectl -n ${NAMESPACE} get pv ${PV_ID} -o yaml | \
yq eval '.spec.nodeAffinity.required.nodeSelectorTerms[0].matchExpressions[] | select(.key == "failure-domain.beta.kubernetes.io/zone").values[0]')

## Skaff versjonen til Loki helm chartet
CHART_VERSION=$(helm -n monitoring ls | grep loki | xargs | cut -d' ' -f9 | cut -d'-' -f2)

echo "Sett nodeSelector til: ${PV_REGION}"

## Hent ut eksisterende config for Loki
helm -n ${NAMESPACE} get values loki > loki-values.yaml

# Finn feltet nodeSelector og legg til failure-domain.beta.kubernetes.io/zone: RELEVANT_REGION
# eksempel:
#
#USER-SUPPLIED VALUES:
#persistence:
#  accessModes:
#  - ReadWriteOnce
#  annotations: {}
#  enabled: true
#  existingClaim: loki
#  selector:
#    matchLabels:
#      app.kubernetes.io/name: loki
#  size: 10Gi
#  subPath: data
#nodeSelector:
#  failure-domain.beta.kubernetes.io/zone: eu-west-1c

# ------------------------------------------------------------------------------------------
### STEP 2
# ------------------------------------------------------------------------------------------

# Kjør følgende kommando for å oppdatere verdier
#helm -n ${NAMESPACE} upgrade loki grafana/loki --version="${CHART_VERSION}" -f loki-values.yaml
